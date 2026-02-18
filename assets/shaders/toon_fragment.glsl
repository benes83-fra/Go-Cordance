#version 410 core

in vec3 Normal;
in vec3 WorldPos;
in vec2 UV;

out vec4 FragColor;

uniform vec3 viewPos;
uniform vec4 baseColor;

void main() {
    vec3 N = normalize(Normal);
    vec3 L = normalize(vec3(0.4, -1.0, 0.2));

    float NdotL = dot(N, L);

    float levels = 4.0;
    float shade = floor(NdotL * levels) / levels;
    shade = max(shade, 0.0);

    FragColor = vec4(baseColor.rgb * shade, 1.0);
}
