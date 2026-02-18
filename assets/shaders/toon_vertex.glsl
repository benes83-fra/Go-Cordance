#version 410 core

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec2 aUV;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

out vec3 Normal;
out vec3 WorldPos;
out vec2 UV;

void main() {
    vec4 world = model * vec4(aPos, 1.0);
    WorldPos = world.xyz;
    Normal = mat3(model) * aNormal;
    UV = aUV;

    gl_Position = projection * view * world;
}
