#version 330 core
in vec3 Normal;
in vec3 Tangent;
in float TangentW;
out vec4 FragColor;

void main() {
    vec3 N = normalize(Normal);
    vec3 T = normalize(Tangent - N * dot(N, Tangent));
    vec3 B = cross(N, T) * TangentW;

    FragColor = vec4(
        normalize(T) * 0.5 + 0.5,
        1.0
    );
}
