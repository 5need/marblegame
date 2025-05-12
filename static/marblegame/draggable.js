/** @typedef {import("p5")} p5 */

export class Draggable {
  /**
   * @param {p5} s
   * @param {number} x
   * @param {number} y
   * @param {number} w
   * @param {number} h
   */
  constructor(s, x, y, w, h) {
    this.dragging = false; // Is the object being dragged?
    this.hovered = false; // Is the mouse over the ellipse?
    this.x = x;
    this.y = y;
    this.w = w;
    this.h = h;
    this.s = s;
    this.offsetX = 0;
    this.offsetY = 0;
  }

  hover(mouseWorldCoords) {
    // Is mouse over object
    if (
      mouseWorldCoords.x > this.x - this.w / 2 &&
      mouseWorldCoords.x < this.x + this.w / 2 &&
      mouseWorldCoords.y > this.y - this.h / 2 &&
      mouseWorldCoords.y < this.y + this.h / 2
    ) {
      this.hovered = true;
    } else {
      this.hovered = false;
    }
  }

  update(mouseWorldCoords) {
    // Adjust location if being dragged
    if (this.dragging) {
      this.x = mouseWorldCoords.x + this.offsetX;
      this.y = mouseWorldCoords.y + this.offsetY;
    }
  }

  show(screenCoords) {
    this.s.stroke(0);
    // Different fill based on state
    if (this.dragging) {
      this.s.fill(50);
    } else if (this.hovered) {
      this.s.fill(100);
    } else {
      this.s.fill(175, 200);
    }
    this.s.rectMode(this.s.CENTER);
    this.s.rect(screenCoords.x, screenCoords.y, this.w, this.h);
  }

  pressed(mouseWorldCoords) {
    // Did I click on the rectangle?
    if (
      mouseWorldCoords.x > this.x - this.w / 2 &&
      mouseWorldCoords.x < this.x + this.w / 2 &&
      mouseWorldCoords.y > this.y - this.h / 2 &&
      mouseWorldCoords.y < this.y + this.h / 2
    ) {
      this.dragging = true;
      // If so, keep track of relative location of click to corner of rectangle
      this.offsetX = this.x - mouseWorldCoords.x;
      this.offsetY = this.y - mouseWorldCoords.y;
    }
  }

  released() {
    // Quit dragging
    this.dragging = false;
  }
}
