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

in vec3 Normal;
in vec3 WorldPos;
in vec2 UV;

out vec4 FragColor;

void main() {
    if (materialType != 2) {
        FragColor = BaseColor;
        return;
    }

    vec3 N = normalize(Normal);
    vec3 L = normalize(vec3(0.4, -1.0, 0.3));

    float NdotL = dot(N, L);

    float levels = 4.0;
    float shade = floor(NdotL * levels) / levels;
    shade = max(shade, 0.0);

    FragColor = vec4(BaseColor.rgb * shade, BaseColor.a);
}
