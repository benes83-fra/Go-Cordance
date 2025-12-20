#version 330 core
in vec3 FragPos;
in vec3 Normal;
in vec2 TexCoord;

out vec4 FragColor;

uniform vec4 BaseColor;
uniform float matAmbient;
uniform float matDiffuse;
uniform float matSpecular;
uniform float matShininess;
uniform vec3 lightDir;
uniform vec3 viewPos;
uniform sampler2D diffuseTex;
uniform bool useTexture;


void main() {
    vec3 norm = normalize(Normal);
    vec3 light = normalize(-lightDir);

    // Ambient
    vec3 ambient = matAmbient * vec3(BaseColor);

    // Diffuse
    float diff = max(dot(norm, light), 0.0);
    vec3 diffuse = matDiffuse * diff * vec3(BaseColor);

    // Specular
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-light, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), matShininess);
    vec3 specular = matSpecular * spec * vec3(1.0);

    vec3 lighting = ambient + diffuse + specular;
  
    vec4 base = BaseColor;

    if (useTexture) {
        base = texture(diffuseTex, TexCoord);
    }

    FragColor = vec4(base.rgb * lighting, base.a);
}
