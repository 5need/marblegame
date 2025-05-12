precision mediump float;

varying vec2 vTexCoord;
uniform sampler2D uTexture;
uniform float uTime;
uniform vec3 uLightDir;
uniform vec4 uQuaternion;
uniform float uOpacity;

const float PI = 3.141592653589793238462643383;

// Rotate on the X axis
mat4 rotationX(float angle) {
    return mat4(
        1., 0., 0., 0.,
        0., cos(angle), -sin(angle), 0.,
        0., sin(angle), cos(angle), 0.,
        0., 0., 0., 1.);
}

// Rotate on the Y axis
mat4 rotationY(float angle) {
    return mat4(
        cos(angle), 0., sin(angle), 0.,
        0., 1.0, 0., 0.,
        -sin(angle), 0., cos(angle), 0.,
        0., 0., 0., 1.);
}

// Rotate on the Z axis
mat4 rotationZ(float angle) {
    return mat4(
        cos(angle), -sin(angle), 0., 0.,
        sin(angle), cos(angle), 0., 0.,
        0., 0., 1., 0.,
        0., 0., 0., 1.);
}

vec3 rotateByQuaternion(vec3 v, vec4 q) {
    vec3 t = 2.0 * cross(q.xyz, v);
    return v + q.w * t + cross(q.xyz, t);
}

void main() {
    // Adjustable parameters
    float shininess = 400.0; // Controls the sharpness of specular highlights
    float rimPower = 4.0; // Higher values make the rim light sharper
    vec3 rimColor = vec3(0.4); // Color of the rim light
    vec3 specularColor = vec3(0.0); // Specular highlight color
    vec3 diffuseColor = vec3(0.9); // Diffuse light multiplier
    vec3 viewDir = normalize(vec3(0.0, 0.0, 1.0)); // Viewer direction

    // Normalize light direction
    vec3 lightDir = normalize(uLightDir);

    // Convert texture coordinates to normalized sphere space
    vec2 uv = (vTexCoord.xy - 0.5) * 2.0;
    float radius = length(uv);

    // Compute normal from 2D coordinates mapped to a sphere
    vec3 normal = vec3(uv.x, uv.y, sqrt(1.0 - uv.x * uv.x - uv.y * uv.y));

    // Lambertian diffuse lighting
    float ndotl = max(0.8, dot(normal, lightDir));
    vec3 diffuse = diffuseColor * ndotl;

    // Compute specular reflection using Blinn-Phong
    vec3 halfVector = normalize(lightDir + viewDir);
    float ndoth = max(0.0, dot(normal, halfVector));
    float specular = pow(ndoth, shininess);
    vec3 specularHighlight = specularColor * specular;

    // Rim lighting calculation
    float ndotv = dot(normal, viewDir);
    float rim = pow(1.0 - ndotv, rimPower);
    vec3 rimLighting = rimColor * rim;

    vec3 rotatedNormal = rotateByQuaternion(normal, uQuaternion);

    float theta = atan(rotatedNormal.x, rotatedNormal.z); // Azimuthal angle (-π to π)
    float phi = asin(rotatedNormal.y); // Elevation angle (-π/2 to π/2)

    // // Convert rotated normal to spherical coordinates
    // float theta = atan(vert.x, vert.z); // Azimuthal angle (-π to π)
    // float phi = asin(vert.y); // Elevation angle (-π/2 to π/2)

    // Map spherical coordinates to 2D texture coordinates
    vec2 texCoords = fract(vec2(0.5 + theta / (1.0 * PI), phi / PI + 0.5));
    // vec2 texCoords = vec2(0.5 + 2.0 * theta / (2.0 * PI), 0.5 + 2.0 * phi / PI);

    // Sample texture
    vec3 texColor = texture2D(uTexture, texCoords).xyz;

    // Combine diffuse, specular, and rim components
    vec3 lightingColor = diffuse + specularHighlight + rimLighting;

    // float intensity = (lightingColor.r + lightingColor.g + lightingColor.b) / 3.0;

    // float levels = 7.0;
    // float level = floor(intensity * levels) / levels;
    //
    // lightingColor = vec3(level);
    // lightingColor = min(lightingColor, vec3(1.2));

    vec3 finalColor = texColor * lightingColor;

    vec4 inside = vec4(finalColor, 1.0) * smoothstep(1.0, 0.97, radius);

    vec4 asdf = vec4(inside.rgb, inside.a);

    gl_FragColor = asdf * uOpacity;
}
