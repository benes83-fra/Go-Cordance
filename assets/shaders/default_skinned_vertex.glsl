#version 330 core
layout(location = 0) in vec3 position;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 texcoord;
layout(location = 3) in vec4 aTangent;
layout(location = 4) in uvec4 aJoints;
layout(location = 5) in vec4 aWeights;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 lightSpaceMatrix;

// joint * inverseBind
uniform mat4 uJointMatrices[128];

out vec4 LightSpacePos;
out vec3 FragPos;
out vec3 Normal;
out vec3 Tangent;
out float TangentW;
out vec2 TexCoord;

void main()
{
    // Skin matrix
    mat4 skinMat =
        aWeights.x * uJointMatrices[aJoints.x] +
        aWeights.y * uJointMatrices[aJoints.y] +
        aWeights.z * uJointMatrices[aJoints.z] +
        aWeights.w * uJointMatrices[aJoints.w];

    // Combined transform
    mat4 modelSkin = model * skinMat;

    vec4 worldPos = modelSkin * vec4(position, 1.0);
    FragPos       = worldPos.xyz;
    LightSpacePos = lightSpaceMatrix * worldPos;

    mat3 normalMatrix = mat3(transpose(inverse(modelSkin)));
    Normal  = normalMatrix * normal;
    Tangent = normalMatrix * aTangent.xyz;
    TangentW = aTangent.w;

    TexCoord = texcoord;
    gl_Position = projection * view * worldPos;
}
