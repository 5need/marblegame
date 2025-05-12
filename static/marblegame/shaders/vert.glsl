attribute vec3 aPosition;
attribute vec2 aTexCoord;

varying vec2 vTexCoord;

void main() {
    gl_Position = vec4(aPosition, 1.0); // Convert to clip space
    vTexCoord = aTexCoord; // Pass texture coordinates to fragment shader
}
