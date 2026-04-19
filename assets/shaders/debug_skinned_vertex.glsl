#version 330 core
layout(location=0) in vec3 position;
layout(location=4) in uvec4 aJoints;
layout(location=5) in vec4 aWeights;

out vec4 vJointColor;
out vec4 vWeightColor;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main() {
    // encode joint indices into color channels (scaled)
    vJointColor = vec4(
        float(aJoints.x) / 255.0,
        float(aJoints.y) / 255.0,
        float(aJoints.z) / 255.0,
        float(aJoints.w) / 255.0
    );

    // encode weights directly
    vWeightColor = aWeights;

    gl_Position = projection * view * model * vec4(position, 1.0);
}
