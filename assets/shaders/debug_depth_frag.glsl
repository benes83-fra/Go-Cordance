#version 330 core
in vec2 TexCoord;
out vec4 FragColor;
uniform sampler2D depthTex;
void main() {
    float d = texture(depthTex, TexCoord).r;
    FragColor = vec4(vec3(d), 1.0);
}
