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

uniform sampler2D diffuseTex;
uniform bool useTexture;

uniform sampler2D normalMap;
uniform bool useNormalMap;
uniform bool flipNormalGreen; // toggle green channel flip for DX vs GL

// Debug mode selector
// 0 = final shaded result
// 1 = normal map raw RGB
// 2 = visualize Tangent (rgb)
// 3 = visualize Bitangent (rgb)
// 4 = visualize Normal (rgb)
// 5 = visualize TangentW (grayscale)
// 6 = visualize UV (rg)
uniform int showMode;

void main() {
    // base geometric normal
    vec3 N = normalize(Normal);

    // orthogonalize tangent against normal
    vec3 T = normalize(Tangent - N * dot(N, Tangent));
    vec3 B = cross(N, T) * TangentW;

    // sample normal map raw
    vec3 nm = vec3(0.5, 0.5, 1.0);
    if (useNormalMap) {
        nm = texture(normalMap, TexCoord).rgb;
        if (flipNormalGreen) nm.g = 1.0 - nm.g;
    }

    // debug outputs
    if (showMode == 1) {
        // show normal map raw
        FragColor = vec4(nm, 1.0);
        return;
    }
    if (showMode == 2) {
        // show tangent direction in world space
        FragColor = vec4(normalize(T) * 0.5 + 0.5, 1.0);
        return;
    }
    if (showMode == 3) {
        // show bitangent direction in world space
        FragColor = vec4(normalize(B) * 0.5 + 0.5, 1.0);
        return;
    }
    if (showMode == 4) {
        // show geometric normal
        FragColor = vec4(normalize(N) * 0.5 + 0.5, 1.0);
        return;
    }
    if (showMode == 5) {
        // show tangent handedness
        float w = (TangentW * 0.5) + 0.5;
        FragColor = vec4(vec3(w), 1.0);
        return;
    }
    if (showMode == 6) {
        // show UVs
        FragColor = vec4(TexCoord, 0.0, 1.0);
        return;
    }

    // final shaded result path
    vec3 finalNormal = N;
    if (useNormalMap) {
        vec3 n = nm * 2.0 - 1.0;
        finalNormal = normalize(mat3(T, B, N) * n);
    }

    vec3 light = normalize(-lightDir);
    vec3 viewDir = normalize(viewPos - FragPos);

    vec3 ambient = matAmbient * vec3(BaseColor);
    float diff = max(dot(finalNormal, light), 0.0);
    vec3 diffuse = matDiffuse * diff * vec3(BaseColor);

    vec3 reflectDir = reflect(-light, finalNormal);
    float spec = 0.0;
    if (diff > 0.0) {
        spec = pow(max(dot(viewDir, reflectDir), 0.0), matShininess);
    }
    vec3 specular = matSpecular * spec * vec3(1.0);

    vec3 lighting = ambient + diffuse + specular;

    vec4 base = BaseColor;
    if (useTexture) base = texture(diffuseTex, TexCoord);

    FragColor = vec4(base.rgb * lighting, base.a);
}
