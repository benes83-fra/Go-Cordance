// debug_fragment.glsl
#version 330 core
out vec4 FragColor;
uniform vec4 debugColor;
void main() {
    FragColor = debugColor;
}