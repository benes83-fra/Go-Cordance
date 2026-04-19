#version 330 core
in vec4 vJointColor;
in vec4 vWeightColor;
out vec4 FragColor;

void main() {
    // show joints in RGB and weights in alpha (or swap)
    FragColor = vec4(vJointColor.rgb, clamp(vWeightColor.x + vWeightColor.y + vWeightColor.z + vWeightColor.w, 0.0, 1.0));
}
