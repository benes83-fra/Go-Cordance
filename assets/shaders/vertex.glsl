#version 330 core
layout(location = 0) in vec3 position;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 texcoord;
layout(location = 3) in vec4 aTangent;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
out vec4 LightSpacePos; 
uniform mat4 lightSpaceMatrix;
out vec3 FragPos;
out vec3 Normal;
out vec3 Tangent;
out float TangentW;
out vec2 TexCoord;
void main() {
    
    vec4 worldPos = model * vec4(position, 1.0);
    FragPos = worldPos.xyz;
    LightSpacePos = lightSpaceMatrix * worldPos;
    Normal  = mat3(transpose(inverse(model))) * normal;
    Tangent = mat3(transpose(inverse(model))) * aTangent.xyz;
    TangentW = aTangent.w;
    TexCoord = texcoord;
    gl_Position = projection * view * worldPos;
}
