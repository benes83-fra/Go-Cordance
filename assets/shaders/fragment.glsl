#version 330 core
in vec3 FragPos;
in vec3 Normal;

out vec4 FragColor;

uniform vec4 baseColor;
uniform float matAmbient;
uniform float matDiffuse;
uniform float matSpecular;
uniform float matShininess;
uniform vec3 lightDir;
uniform vec3 viewPos;

void main() {
    vec3 norm = normalize(Normal);
    vec3 light = normalize(-lightDir);

    // Ambient
    vec3 ambient = matAmbient * vec3(baseColor);

    // Diffuse
    float diff = max(dot(norm, light), 0.0);
    vec3 diffuse = matDiffuse * diff * vec3(baseColor);

    // Specular
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-light, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), matShininess);
    vec3 specular = matSpecular * spec * vec3(1.0);

    vec3 result = ambient + diffuse + specular;
    FragColor = vec4(result, baseColor.a);
}
