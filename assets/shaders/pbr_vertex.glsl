#version 410 core

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec2 aUV;
layout(location = 3) in vec3 aTangent;
layout(location = 4) in vec3 aBitangent;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

out VS_OUT {
    vec3 WorldPos;
    vec3 Normal;
    vec2 UV;
} vs_out;

void main() {
    vec4 world = model * vec4(aPos, 1.0);
    vs_out.WorldPos = world.xyz;
    vs_out.Normal = mat3(model) * aNormal;
    vs_out.UV = aUV;

    gl_Position = projection * view * world;
}
