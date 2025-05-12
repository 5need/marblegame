precision mediump float;

varying vec2 pos;

uniform float millis;

const int num_circles = 100;
uniform vec3 circles[num_circles];

void main() {
    float color = 1.;
    for (int i = 0; i < num_circles; i++) {
        float d = length(pos - circles[i].xy) - circles[i].z;
        d = smoothstep(0., 0.005, d);
        color *= d;
    }

    vec4 c2 = vec4(color, color, color, 1.);
    gl_FragColor = c2;
}
