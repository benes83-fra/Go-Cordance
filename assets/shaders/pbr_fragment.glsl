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
uniform sampler2D albedoTex;
uniform bool useTexture;

in VS_OUT {
    vec3 WorldPos;
    vec3 Normal;
    vec2 UV;
} fs_in;

out vec4 FragColor;

void main() {
    vec3 N = normalize(fs_in.Normal);
    vec3 V = normalize(viewPos - fs_in.WorldPos);

    vec3 albedo = BaseColor.rgb;
    if (useTexture)
        albedo = texture(albedoTex, fs_in.UV).rgb;

    float m = metallic;
    float r = roughness;

    vec3 L = normalize(vec3(0.3, -1.0, 0.2));
    vec3 H = normalize(V + L);

    float NdotL = max(dot(N, L), 0.0);
    float spec = pow(max(dot(N, H), 0.0), mix(8.0, 128.0, 1.0 - r));

    vec3 color = albedo * NdotL + spec * (1.0 - r);

    FragColor = vec4(color, 1.0);
}
