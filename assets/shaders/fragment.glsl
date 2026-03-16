#version 330 core
#define MAX_LIGHTS 8

in vec3 FragPos;
in vec3 Normal;
in vec4 LightSpacePos;
in vec3 Tangent;
in float TangentW;
in vec2 TexCoord;

out vec4 FragColor;

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
};

uniform vec3 viewPos;

uniform int   lightCount;
uniform vec3  lightDir[MAX_LIGHTS];
uniform vec3  lightColor[MAX_LIGHTS];
uniform float lightIntensity[MAX_LIGHTS];
uniform vec3  lightPos[MAX_LIGHTS];
uniform float lightRange[MAX_LIGHTS];
uniform float lightAngle[MAX_LIGHTS];
uniform int   lightType[MAX_LIGHTS];

uniform sampler2D shadowMap;
uniform mat4 lightSpaceMatrix;
uniform vec2 uShadowMapSize;
uniform int  shadowLightIndex;

// textures
uniform sampler2D diffuseTex;
uniform bool      useTexture;

uniform sampler2D normalMap;
uniform bool      useNormalMap;

uniform sampler2D occlusionMap;
uniform sampler2D metallicRoughnessMap;
uniform bool      useOcclusionMap;
uniform bool      useMetallicRoughnessMap;

// UV controls (match Go: LocUVScaleBase / LocUVOffsetBase)
uniform vec2 uvScaleBase;
uniform vec2 uvOffsetBase;

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

void main() {
    // This shader is Blinn/Phong only
    if (materialType != 0) {
        FragColor = BaseColor;
        return;
    }

    // TBN
    vec3 N = normalize(Normal);
    vec3 T = normalize(Tangent - N * dot(N, Tangent));
    vec3 B = cross(N, T) * TangentW;
    mat3 TBN = mat3(T, B, N);

    vec3 finalNormal = N;
    if (useNormalMap) {
        vec3 n = texture(normalMap, TexCoord).rgb;
        n = n * 2.0 - 1.0;
        finalNormal = normalize(TBN * n);
    }

    vec3 viewDir = normalize(viewPos - FragPos);

    // UVs
    vec2 baseUV = TexCoord * uvScaleBase + uvOffsetBase;

    // Base color
    vec4 base = BaseColor;
    if (useTexture) {
        base = texture(diffuseTex, baseUV);
    }

    // Occlusion
    float ao = 1.0;
    if (useOcclusionMap) {
        ao = texture(occlusionMap, baseUV).r;
    }

    // Metallic/Roughness (not really used in Blinn/Phong, but read safely)
    vec2 mr = vec2(metallic, roughness);
    if (useMetallicRoughnessMap) {
        mr = texture(metallicRoughnessMap, baseUV).bg;
    }
    float metallicVal  = mr.x;
    float roughnessVal = mr.y;

    // Ambient
    vec3 ambient = matAmbient * vec3(base) * ao;

    vec3 lighting = ambient;

    // Shadow-space position
    vec4 lightSpacePos = lightSpaceMatrix * vec4(FragPos, 1.0);

    for (int i = 0; i < lightCount; i++) {
        vec3 L;
        float attenuation = 1.0;
        float shadow = 1.0;

        if (shadowLightIndex >= 0 &&
            i == shadowLightIndex &&
            (lightType[i] == 0 || lightType[i] == 2)) {
            shadow = computeShadowPCF(lightSpacePos, finalNormal, -lightDir[i]);
        }

        if (lightType[i] == 0) {
            L = normalize(-lightDir[i]);
        } else {
            vec3 toLight = lightPos[i] - FragPos;
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

        float diff = max(dot(finalNormal, L), 0.0);
        vec3 diffuse = matDiffuse * diff * lightColor[i] * lightIntensity[i];

        vec3 H = normalize(L + viewDir);
        float spec = pow(max(dot(finalNormal, H), 0.0), matShininess);
        vec3 specular = matSpecular * spec * lightColor[i] * lightIntensity[i];

        float shadowFactor = 1.0;
        if (i == shadowLightIndex && (lightType[i] == 0 || lightType[i] == 2)) {
            shadowFactor = 1.0 - shadow;
        }

        lighting += (diffuse + specular) * attenuation * shadowFactor;
    }

    FragColor = vec4(base.rgb * lighting, base.a);
}
