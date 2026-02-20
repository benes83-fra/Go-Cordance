#version 410 core
layout(std140) uniform MaterialBlock {
    vec4  uBaseColor;
    float uAmbient;
    float uDiffuse;
    float uSpecular;
    float uShininess;
};

uniform vec3 viewPos;

in vec3 Normal;
in vec3 WorldPos;
in vec2 UV;

out vec4 FragColor;

void main() {
    vec3 N = normalize(Normal);
    vec3 L = normalize(vec3(0.4, -1.0, 0.2));

    float NdotL = dot(N, L);

    float levels = 4.0;
    float shade = floor(NdotL * levels) / levels;
    shade = max(shade, 0.0);

    FragColor = vec4(uBaseColor.rgb * shade, uBaseColor.a);
}
