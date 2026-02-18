#version 410 core

in VS_OUT {
    vec3 WorldPos;
    vec3 Normal;
    vec2 UV;
} fs_in;

out vec4 FragColor;

uniform vec3 viewPos;

uniform vec4 baseColor;
uniform float metallic;
uniform float roughness;

uniform sampler2D albedoTex;
uniform bool useTexture;

void main() {
    vec3 N = normalize(fs_in.Normal);
    vec3 V = normalize(viewPos - fs_in.WorldPos);

    vec3 albedo = baseColor.rgb;
    if (useTexture)
        albedo = texture(albedoTex, fs_in.UV).rgb;

    float m = metallic;
    float r = roughness;

    // Simple lighting: 1 directional light
    vec3 L = normalize(vec3(0.3, -1.0, 0.2));
    vec3 H = normalize(V + L);

    float NdotL = max(dot(N, L), 0.0);

    // Fake specular
    float spec = pow(max(dot(N, H), 0.0), mix(8.0, 128.0, 1.0 - r));

    vec3 color = albedo * NdotL + spec * (1.0 - r);

    FragColor = vec4(color, 1.0);
}
