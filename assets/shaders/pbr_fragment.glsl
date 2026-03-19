#version 410 core
#define MAX_LIGHTS 8

layout(std140) uniform MaterialBlock {
    vec4 BaseColor;
    float matAmbient;
    float matDiffuse;
    float matSpecular;
    float matShininess;

    float metallic;
    float roughness;
    int   materialType;
    float _pad0;
    vec4 EmissiveColor;
};

uniform vec3 viewPos;

// Albedo
uniform sampler2D albedoTex;
uniform bool useTexture;

// Normal map
uniform sampler2D normalMap;
uniform bool useNormalMap;

// Metallic-Roughness map
uniform sampler2D metallicRoughnessMap;
uniform bool useMetallicRoughnessMap;

// Ambient Occlusion map
uniform sampler2D occlusionMap;
uniform bool useOcclusionMap;

// UV transforms
uniform vec2 uvScaleBase;
uniform vec2 uvOffsetBase;

uniform vec2 uvScaleNormal;
uniform vec2 uvOffsetNormal;

uniform vec2 uvScaleOcclusion;
uniform vec2 uvOffsetOcclusion;

uniform vec2 uvScaleMR;
uniform vec2 uvOffsetMR;

// TexCoordMap indices
uniform int texCoordBase;
uniform int texCoordNormal;
uniform int texCoordOcclusion;
uniform int texCoordMR;

// Lights
uniform int   lightCount;
uniform vec3  lightDir[MAX_LIGHTS];
uniform vec3  lightColor[MAX_LIGHTS];
uniform float lightIntensity[MAX_LIGHTS];
uniform vec3  lightPos[MAX_LIGHTS];
uniform float lightRange[MAX_LIGHTS];
uniform float lightAngle[MAX_LIGHTS];
uniform int   lightType[MAX_LIGHTS];

// Shadows
uniform sampler2D shadowMap;
uniform vec2 uShadowMapSize;
uniform int  shadowLightIndex;
uniform float normalScale;
// Emissive
uniform sampler2D emissiveTex;
uniform bool useEmissiveTex;

// Emissive UV transform + texcoord
uniform vec2 uvScaleEmissive;
uniform vec2 uvOffsetEmissive;
uniform int  texCoordEmissive;

// IBL
uniform samplerCube irradianceMap;      // diffuse
uniform samplerCube prefilteredEnvMap;  // specular (mipmapped)
uniform sampler2D   brdfLUT;            // 2D LUT
uniform bool        useIBL;


in VS_OUT {
    vec3 WorldPos;
    vec3 Normal;
    vec2 UV0;
    vec2 UV1;
    vec3 Tangent;
    vec3 Bitangent;
    vec4 LightSpacePos;
} fs_in;

out vec4 FragColor;

// ------------------------------------------------------------
// UV selection helper
// ------------------------------------------------------------
vec2 selectUV(int setIndex, vec2 uv0, vec2 uv1)
{
    return (setIndex == 1) ? uv1 : uv0;
}

// ------------------------------------------------------------
// Shadow helper
// ------------------------------------------------------------
float computeShadowPCF(vec4 lightSpacePos, vec3 normal, vec3 lightDirWS)
{
    vec3 projCoords = lightSpacePos.xyz / lightSpacePos.w;
    vec2 uv = projCoords.xy * 0.5 + 0.5;
    float currentDepth = projCoords.z * 0.5 + 0.5;

    if (uv.x < 0.0 || uv.x > 1.0 ||
        uv.y < 0.0 || uv.y > 1.0 ||
        currentDepth < 0.0 || currentDepth > 1.0) {
        return 0.0;
    }

    float ndotl = max(dot(normalize(normal), normalize(lightDirWS)), 0.0);
    float bias = mix(0.0005, 0.00005, ndotl);

    vec2 texelSize = 0.7 / uShadowMapSize;

    float shadow = 0.0;
    int samples = 0;
    for (int y = -1; y <= 1; ++y) {
        for (int x = -1; x <= 1; ++x) {
            vec2 offset = vec2(x, y) * texelSize;
            float closestDepth = texture(shadowMap, uv + offset).r;
            shadow += currentDepth - bias > closestDepth ? 1.0 : 0.0;
            samples++;
        }
    }

    return shadow / float(samples);
}

// ------------------------------------------------------------
// PBR helper functions
// ------------------------------------------------------------
float DistributionGGX(vec3 N, vec3 H, float roughness)
{
    float a  = roughness * roughness;
    float a2 = a * a;
    float NdotH  = max(dot(N, H), 0.0);
    float NdotH2 = NdotH * NdotH;

    float denom = (NdotH2 * (a2 - 1.0) + 1.0);
    return a2 / (3.14159 * denom * denom);
}

float GeometrySchlickGGX(float NdotV, float roughness)
{
    float r = roughness + 1.0;
    float k = (r * r) / 8.0;

    return NdotV / (NdotV * (1.0 - k) + k);
}

float GeometrySmith(vec3 N, vec3 V, vec3 L, float roughness)
{
    float NdotV = max(dot(N, V), 0.0);
    float NdotL = max(dot(N, L), 0.0);

    float ggx1 = GeometrySchlickGGX(NdotV, roughness);
    float ggx2 = GeometrySchlickGGX(NdotL, roughness);

    return ggx1 * ggx2;
}

vec3 FresnelSchlick(float cosTheta, vec3 F0)
{
    return F0 + (1.0 - F0) * pow(1.0 - cosTheta, 5.0);
}



// --------------------------------------------------------
// Tone mapping + gamma correction
// --------------------------------------------------------
vec3 ACESFilm(vec3 x)
{
    float a = 2.51;
    float b = 0.03;
    float c = 2.43;
    float d = 0.59;
    float e = 0.14;
    return clamp((x*(a*x+b)) / (x*(c*x+d)+e), 0.0, 1.0);
}




// ------------------------------------------------------------
// Main
// ------------------------------------------------------------
void main()
{
    if (materialType != 1) {
        FragColor = BaseColor;
        return;
    }

    // --------------------------------------------------------
    // Normal + view
    // --------------------------------------------------------
    vec3 N = normalize(fs_in.Normal);
    vec3 V = normalize(viewPos - fs_in.WorldPos);

    // --------------------------------------------------------
    // Normal mapping
    // --------------------------------------------------------
    // Normal mapping
    vec2 uvN = selectUV(texCoordNormal, fs_in.UV0, fs_in.UV1);
    uvN = uvN * uvScaleNormal + uvOffsetNormal;

    if (useNormalMap) {
        vec3 T = normalize(fs_in.Tangent);
        vec3 B = normalize(fs_in.Bitangent);
        mat3 TBN = mat3(T, B, N);

        vec3 n = texture(normalMap, uvN).rgb;
        n = n * 2.0 - 1.0;

        // GLTF normalTexture.scale
        n.xy *= normalScale;

        n = normalize(n);
        N = normalize(TBN * n);
    }


    // --------------------------------------------------------
    // Albedo
    // --------------------------------------------------------
    vec2 uvBase = selectUV(texCoordBase, fs_in.UV0, fs_in.UV1);
    uvBase = uvBase * uvScaleBase + uvOffsetBase;

    vec3 albedo = BaseColor.rgb;
    if (useTexture) {
        albedo = texture(albedoTex, uvBase).rgb;
    }

    // --------------------------------------------------------
    // Metallic-Roughness
    // --------------------------------------------------------
    vec2 uvMR = selectUV(texCoordMR, fs_in.UV0, fs_in.UV1);
    uvMR = uvMR * uvScaleMR + uvOffsetMR;

    vec2 mr = vec2(metallic, roughness);
    if (useMetallicRoughnessMap) {
        vec3 mrTex = texture(metallicRoughnessMap, uvMR).rgb;
        mr = vec2(mrTex.b, mrTex.g);
    }

    float m = mr.x;
    float r = mr.y;

    // --------------------------------------------------------
    // Ambient Occlusion
    // --------------------------------------------------------
    vec2 uvAO = selectUV(texCoordOcclusion, fs_in.UV0, fs_in.UV1);
    uvAO = uvAO * uvScaleOcclusion + uvOffsetOcclusion;

    float ao = 1.0;
    if (useOcclusionMap) {
        ao = texture(occlusionMap, uvAO).r;
    }

    // --------------------------------------------------------
    // Fresnel base reflectance
    // --------------------------------------------------------
    vec3 F0 = mix(vec3(0.04), albedo, m);

    // --------------------------------------------------------
    // Ambient
    // --------------------------------------------------------
    vec3 ambient = albedo * 0.03 * ao;

    vec3 color = ambient;

    // --------------------------------------------------------
    // Per-light loop with shadows
    // --------------------------------------------------------
    for (int i = 0; i < lightCount; ++i) {
        vec3 L;
        float attenuation = 1.0;

        if (lightType[i] == 0) {
            L = normalize(-lightDir[i]);
        } else {
            vec3 toLight = lightPos[i] - fs_in.WorldPos;
            float dist = length(toLight);
            L = normalize(toLight);
            attenuation = 1.0 / (1.0 + (dist / lightRange[i]) * (dist / lightRange[i]));
        }

        if (lightType[i] == 2) {
            float cutoff = cos(radians(lightAngle[i]));
            float spotFactor = dot(L, normalize(-lightDir[i]));
            if (spotFactor < cutoff) {
                continue;
            }
            float spotSmooth = (spotFactor - cutoff) / (1.0 - cutoff);
            attenuation *= spotSmooth;
        }

        float NdotL = max(dot(N, L), 0.0);
        if (NdotL <= 0.0) continue;

        vec3 H = normalize(V + L);

        float D = DistributionGGX(N, H, r);
        float G = GeometrySmith(N, V, L, r);
        vec3  F = FresnelSchlick(max(dot(H, V), 0.0), F0);

        vec3 numerator = D * G * F;
        float denom = 4.0 * max(dot(N, V), 0.0) * NdotL + 0.001;
        vec3 specular = numerator / denom;

        vec3 kd = (1.0 - F) * (1.0 - m);
        vec3 diffuse = kd * albedo / 3.14159;

        float shadowFactor = 1.0;
        if (shadowLightIndex >= 0 &&
            i == shadowLightIndex &&
            (lightType[i] == 0 || lightType[i] == 2)) {
            float shadow = computeShadowPCF(fs_in.LightSpacePos, N, -lightDir[i]);
            shadowFactor = 1.0 - shadow;
        }

        vec3 radiance = lightColor[i] * lightIntensity[i];

        color += (diffuse + specular) * radiance * NdotL * attenuation * shadowFactor;
    }
        // ... end of direct lighting loop ...

    // --------------------------------------------------------
    // Image-Based Lighting (IBL)
    // --------------------------------------------------------
    if (useIBL) {
        // Diffuse IBL (irradiance)
        vec3 irradiance = texture(irradianceMap, N).rgb;
        vec3 F_ibl = FresnelSchlick(max(dot(N, V), 0.0), F0);
        vec3 kd_ibl = (1.0 - F_ibl) * (1.0 - m);
        vec3 diffuseIBL = kd_ibl * irradiance * albedo;

        // Specular IBL (prefiltered env + BRDF LUT)
        vec3 R = reflect(-V, N);
        float mipCount = 5.0; // adjust to your prefiltered env mip levels
        vec3 prefilteredColor = textureLod(prefilteredEnvMap, R, r * mipCount).rgb;

        vec2 brdfSample = texture(brdfLUT, vec2(max(dot(N, V), 0.0), r)).rg;
        vec3 specularIBL = prefilteredColor * (F_ibl * brdfSample.x + brdfSample.y);

        color += diffuseIBL + specularIBL;
    }

    // --------------------------------------------------------
    // Emissive
    // --------------------------------------------------------
    vec3 emissive = vec3(0.0);

    // Base emissive color from material (if you added EmissiveColor)
    emissive += EmissiveColor.rgb;

    // Emissive texture
    vec2 uvEm = selectUV(texCoordEmissive, fs_in.UV0, fs_in.UV1);
    uvEm = uvEm * uvScaleEmissive + uvOffsetEmissive;

    if (useEmissiveTex) {
        emissive += texture(emissiveTex, uvEm).rgb;
    }

    // Add emissive in linear space
    color += emissive;

    // Tone map (ACES)
    color = ACESFilm(color);

    // Gamma correction (linear → sRGB)
    color = pow(color, vec3(1.0/2.2));

    FragColor = vec4(color, 1.0);

}
