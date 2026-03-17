#version 410 core

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

in VS_OUT {
    vec3 WorldPos;
    vec3 Normal;
    vec2 UV;
    vec3 Tangent;
    vec3 Bitangent;
} fs_in;

out vec4 FragColor;

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
    // Base inputs
    // --------------------------------------------------------
    vec3 N = normalize(fs_in.Normal);
    vec3 V = normalize(viewPos - fs_in.WorldPos);

    // --------------------------------------------------------
    // Normal mapping
    // --------------------------------------------------------
    if (useNormalMap) {
        vec3 T = normalize(fs_in.Tangent);
        vec3 B = normalize(fs_in.Bitangent);
        mat3 TBN = mat3(T, B, N);

        vec3 n = texture(normalMap, fs_in.UV).rgb;
        n = n * 2.0 - 1.0;

        N = normalize(TBN * n);
    }

    // --------------------------------------------------------
    // Albedo
    // --------------------------------------------------------
    vec3 albedo = BaseColor.rgb;
    if (useTexture) {
        albedo = texture(albedoTex, fs_in.UV).rgb;
    }

    // --------------------------------------------------------
    // Metallic-Roughness
    // --------------------------------------------------------
    vec2 mr = vec2(metallic, roughness);

    if (useMetallicRoughnessMap) {
        vec3 mrTex = texture(metallicRoughnessMap, fs_in.UV).rgb;
        mr = vec2(mrTex.b, mrTex.g); // (metallic, roughness)
    }

    float m = mr.x;
    float r = mr.y;

    // --------------------------------------------------------
    // Ambient Occlusion
    // --------------------------------------------------------
    float ao = 1.0;
    if (useOcclusionMap) {
        ao = texture(occlusionMap, fs_in.UV).r;
    }

    // --------------------------------------------------------
    // Light
    // --------------------------------------------------------
    vec3 L = normalize(vec3(0.3, -1.0, 0.2));
    vec3 H = normalize(V + L);

    float NdotL = max(dot(N, L), 0.0);

    // --------------------------------------------------------
    // Fresnel reflectance
    // --------------------------------------------------------
    vec3 F0 = mix(vec3(0.04), albedo, m);

    // --------------------------------------------------------
    // BRDF
    // --------------------------------------------------------
    float D = DistributionGGX(N, H, r);
    float G = GeometrySmith(N, V, L, r);
    vec3  F = FresnelSchlick(max(dot(H, V), 0.0), F0);

    vec3 numerator = D * G * F;
    float denom = 4.0 * max(dot(N, V), 0.0) * NdotL + 0.001;
    vec3 specular = numerator / denom;

    vec3 kd = (1.0 - F) * (1.0 - m);
    vec3 diffuse = kd * albedo / 3.14159;

    // --------------------------------------------------------
    // Ambient + Direct lighting
    // --------------------------------------------------------
    vec3 ambient = albedo * 0.03 * ao;
    vec3 color = ambient + (diffuse + specular) * NdotL;

    FragColor = vec4(color, 1.0);
}
