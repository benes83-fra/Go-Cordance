#version 330 core
in vec3 Normal;
out vec4 FragColor;
void main() {
    vec3 n = normalize(Normal) * 0.5 + 0.5;
    FragColor = vec4(n, 1.0);
}
