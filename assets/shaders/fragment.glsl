#version 330 core

in vec3 FragPos;
in vec3 Normal;
in vec3 Tangent;
in float TangentW;
in vec2 TexCoord;

out vec4 FragColor;

uniform vec4 BaseColor;
uniform float matAmbient;
uniform float matDiffuse;
uniform float matSpecular;
uniform float matShininess;
uniform vec3 lightDir;
uniform vec3 viewPos;

// diffuse texture
uniform sampler2D diffuseTex;
uniform bool useTexture;

// normal map (optional)
uniform sampler2D normalMap;
uniform bool useNormalMap;

void main() {
    // Reconstruct TBN
    vec3 N = normalize(Normal);
    // Orthogonalize tangent against normal
    vec3 T = normalize(Tangent - N * dot(N, Tangent));
    vec3 B = cross(N, T) * TangentW;
    mat3 TBN = mat3(T, B, N);

    // Sample normal map in tangent space if available
    vec3 finalNormal = N;
    if (useNormalMap) {
        vec3 n = texture(normalMap, TexCoord).rgb;
        n = n * 2.0 - 1.0; // map from [0,1] to [-1,1]
        finalNormal = normalize(TBN * n);
    }

    vec3 light = normalize(-lightDir);
    vec3 viewDir = normalize(viewPos - FragPos);

    // Ambient
    vec3 ambient = matAmbient * vec3(BaseColor);

    // Diffuse (use finalNormal)
    float diff = max(dot(finalNormal, light), 0.0);
    vec3 diffuse = matDiffuse * diff * vec3(BaseColor);

    // Specular (use finalNormal)
    vec3 reflectDir = reflect(-light, finalNormal);
    float spec = 0.0;
    if (diff > 0.0) {
        spec = pow(max(dot(viewDir, reflectDir), 0.0), matShininess);
    }
    vec3 specular = matSpecular * spec * vec3(1.0);

    vec3 lighting = ambient + diffuse + specular;

    // Base color from material or diffuse texture
    vec4 base = BaseColor;
    if (useTexture) {
        base = texture(diffuseTex, TexCoord);
    }

    // Compose final color; keep alpha from base
    FragColor = vec4(base.rgb * lighting, base.a);
}
