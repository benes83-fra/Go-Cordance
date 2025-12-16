#version 410
in vec3 vp;
in vec3 color;
out vec3 fragColor;
void main() {
    fragColor = color;
    gl_Position = vec4(vp, 1.0);
}
