#version 330 core
#define MAX_LIGHTS 8

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

uniform vec3 viewPos;


uniform int lightCount;
uniform vec3 lightDir[MAX_LIGHTS];
uniform vec3 lightColor[MAX_LIGHTS];
uniform float lightIntensity[MAX_LIGHTS];
uniform vec3 lightPos[MAX_LIGHTS];
uniform float lightRange[MAX_LIGHTS];
uniform float lightAngle[MAX_LIGHTS]; // degrees
uniform int lightType[MAX_LIGHTS];    // 0=dir, 1=point, 2=spot

uniform sampler2D shadowMap;
uniform mat4 lightSpaceMatrix;
uniform vec2 uShadowMapSize;

// diffuse texture
uniform sampler2D diffuseTex;
uniform bool useTexture;

// normal map (optional)
uniform sampler2D normalMap;
uniform bool useNormalMap;
    // (width, height) of the shadow map


float computeShadowPCF(vec4 lightSpacePos)
{
    // perspective divide
    vec3 projCoords = lightSpacePos.xyz / lightSpacePos.w;

    // transform from [-1,1] to [0,1]
    vec2 uv = projCoords.xy * 0.5 + 0.5;
    float currentDepth = projCoords.z * 0.5 + 0.5;

    // outside the light frustum -> treat as lit (no shadow)
    if (uv.x < 0.0 || uv.x > 1.0 ||
        uv.y < 0.0 || uv.y > 1.0 ||
        currentDepth < 0.0 || currentDepth > 1.0) {
        return 0.0;
    }

    // basic bias to reduce shadow acne (tweak as needed)
    float bias = 0.0008;

    // texel size in UV space
    vec2 texelSize = 1.0 / uShadowMapSize;

    // 3x3 kernel around current uv
    float shadow = 0.0;
    int samples = 0;

    for (int y = -1; y <= 1; ++y) {
        for (int x = -1; x <= 1; ++x) {
            vec2 offset = vec2(x, y) * texelSize;
            float closestDepth = texture(shadowMap, uv + offset).r;
            shadow += currentDepth - bias > closestDepth ? 1.0 : 0.0;
            samples++;
        }
    }

    // average result: 0.0 = fully lit, 1.0 = fully shadowed
    return shadow / float(samples);
}


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

    vec3 viewDir = normalize(viewPos - FragPos);

    // Ambient stays the same
    vec3 ambient = matAmbient * vec3(BaseColor);

    // Accumulate all lights
    vec3 lighting = ambient;
    // --- SHADOW CALCULATION ---

    // Transform fragment position into light space
    vec4 lightSpacePos = lightSpaceMatrix * vec4(FragPos, 1.0);

    // Perspective divide
    vec3 projCoords = lightSpacePos.xyz / lightSpacePos.w;

    // Transform from [-1,1] to [0,1]
    projCoords = projCoords * 0.5 + 0.5;

    // Read depth from shadow map
  /*  float closestDepth = texture(shadowMap, projCoords.xy).r;

    // Current fragment depth in light space
    float currentDepth = projCoords.z;

    // Shadow test (with small bias)
    float bias = 0.005;*/
    float shadow = computeShadowPCF(lightSpacePos);



    for (int i = 0; i < lightCount; i++) {

        vec3 L;
        float attenuation = 1.0;

        if (lightType[i] == 0) {
            // -------------------------
            // Directional Light
            // -------------------------
            L = normalize(-lightDir[i]);
        }
        else {
            // -------------------------
            // Point or Spot Light
            // -------------------------
            vec3 toLight = lightPos[i] - FragPos;
            float dist = length(toLight);
            L = normalize(toLight);

            // Attenuation (inverse square falloff)
            attenuation = 1.0 / (1.0 + (dist / lightRange[i]) * (dist / lightRange[i]));
        }

        // -------------------------
        // Spotlight cone falloff
        // -------------------------
        if (lightType[i] == 2) { // SPOTLIGHT
            // Angle in degrees → cosine
            float cutoff = cos(radians(lightAngle[i]));
            float spotFactor = dot(L, normalize(-lightDir[i]));

            if (spotFactor < cutoff) {
                // Outside the cone → no contribution
                continue;
            }

            // Smooth falloff inside the cone
            float spotSmooth = (spotFactor - cutoff) / (1.0 - cutoff);
            attenuation *= spotSmooth;

        }

        // -------------------------
        // Diffuse
        // -------------------------
        float diff = max(dot(finalNormal, L), 0.0);
        vec3 diffuse = matDiffuse * diff * lightColor[i] * lightIntensity[i];

        // -------------------------
        // Specular (Blinn-Phong)
        // -------------------------
        vec3 H = normalize(L + viewDir);
        float spec = pow(max(dot(finalNormal, H), 0.0), matShininess);
        vec3 specular = matSpecular * spec * lightColor[i] * lightIntensity[i];

        // -------------------------
        // Accumulate with attenuation
        // -------------------------
     lighting += (diffuse + specular) * attenuation * (1.0 - shadow);

}





    // Base color from material or diffuse texture
    vec4 base = BaseColor;
    if (useTexture) {
        base = texture(diffuseTex, TexCoord);
    }

    // Compose final color; keep alpha from base
    FragColor = vec4(base.rgb * lighting, base.a);
}



