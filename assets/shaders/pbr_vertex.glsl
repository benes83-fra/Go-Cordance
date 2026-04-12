#version 410 core

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec2 aUV;
layout(location = 3) in vec3 aTangent;
layout(location = 4) in vec3 aBitangent;
layout(location = 5) in uvec4 aJoints;
layout(location = 6) in vec4 aWeights;


uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 lightSpaceMatrix;

out VS_OUT {
    vec3 WorldPos;
    vec3 Normal;
    vec2 UV0;
    vec2 UV1;
    vec3 Tangent;
    vec3 Bitangent;
    vec4 LightSpacePos;
} vs_out;

void main() {
    vec4 world = model * vec4(aPos, 1.0);
    vs_out.WorldPos = world.xyz;

    mat3 normalMatrix = mat3(model);
    vs_out.Normal    = normalize(normalMatrix * aNormal);
    vs_out.Tangent   = normalize(normalMatrix * aTangent);
    vs_out.Bitangent = normalize(normalMatrix * aBitangent);

    vs_out.UV0 = aUV;
    vs_out.UV1 = aUV; // same for now; TexCoordMap can still switch

    vs_out.LightSpacePos = lightSpaceMatrix * world;

    gl_Position = projection * view * world;
}
