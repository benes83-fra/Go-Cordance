#version 330 core
in vec3 FragPos;
uniform vec3 viewPos;
out vec4 FragColor;

void main() {
    float d = length(viewPos - FragPos) / 50.0;
    FragColor = vec4(vec3(d), 1.0);
}
