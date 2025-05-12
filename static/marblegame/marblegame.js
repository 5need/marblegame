/** @typedef {import("p5")} p5 */
/** @typedef {import("p5").Font} p5.Font */
/** @typedef {import("p5").Image} p5.Image */
/** @typedef {import("p5").SoundFile} p5.SoundFile */
/** @typedef {import("p5").Shader} p5.Shader */

/** @import * as M from './marblegametypes' */
/** @import * as H from './howler' */

import "./p5.min.js";
import "./p5.sound.min.js";

import "./howler.js";

import * as draggable from "./draggable.js";

function getCookie(name) {
  const cookies = document.cookie.split("; ");
  for (let cookie of cookies) {
    const [key, value] = cookie.split("=");
    if (key === name) {
      return decodeURIComponent(value);
    }
  }
  return null;
}

/** @type {string} */
const userToken = getCookie("userToken");

/** @type {number} */
let selectedInventorySlot = 0;

/** @type {H.Howl} */
let hitSound;
/** @type {boolean} */
let canPlayAudio = false;

/** @type {p5.Image} */
let cursorImg;

/** @type {p5.Image} */
let marbleTextureImg;

/** @type {p5.Font} */
let font;

/** @type {p5.Shader} */
let shaderProgram;

/** @type HTMLElement */
let cursorForm;
/** @type HTMLElement */
let gameForm;

/** @type {M.MarbleGame} */
let game;

/** @typedef {{userToken: string, x: number, y: number}} CursorPosition */
/** @type {CursorPosition[]} */
let opponentCursorPositionHistory = [];

let mouseWorldCoords = {
  x: 0,
  y: 0,
};

let mouseScreenCoords = {
  x: 0,
  y: 0,
};

let mousePressedWorldCoords = {
  x: 0,
  y: 0,
};

let worldViewOffset = {
  x: 0,
  y: 0,
};

/** @type {Boolean} */
let isMyTurn = false;

let frameIndex = -1;

let marblePlaceholder;
let powerPlaceholder;

/**
 * @param {p5} s
 */
function mySketch(s) {
  s.preload = function () {
    cursorImg = s.loadImage("/images/pointer.svg");
    marbleTextureImg = s.loadImage("/marblegame/marble_texture.jpg");
    // marbleTextureImg = s.loadImage("/marblegame/test_texture.png");
    hitSound = new Howl({
      src: ["/audio/collision.ogg"],
      onunlock: function () {
        canPlayAudio = true;
      },
    });
    font = s.loadFont("/fonts/Inconsolata.otf");
    shaderProgram = s.loadShader(
      "/marblegame/shaders/vert.glsl",
      "/marblegame/shaders/frag.glsl",
    );
  };
  s.setup = function () {
    const canvas = s.createCanvas(
      window.innerWidth,
      window.innerHeight,
      s.WEBGL,
    );
    canvas.parent("game-canvas");
    marblePlaceholder = new draggable.Draggable(s, 0, 0, 200, 200);
    powerPlaceholder = new draggable.Draggable(s, 50, 300, 20, 20);
    s.angleMode(s.DEGREES);
    s.textFont(font);
    s.background(200);
    s.frameRate(60);
    shaderProgram.setUniform("uTexture", marbleTextureImg);

    cursorForm = window.document.getElementById("cursor-form");
    gameForm = window.document.getElementById("game-form");
  };
  s.windowResized = function () {
    s.resizeCanvas(window.innerWidth, window.innerHeight);
  };

  s.draw = function () {
    s.clear();
    s.ortho();
    s.background("#1e1e2e");
    s.translate(-s.width / 2, -s.height / 2);

    const readyToShowGame = !!game && game.frames;

    if (readyToShowGame) {
      mouseWorldCoords = screenCoordsToWorldCoords(s, s.mouseX, s.mouseY);
      mouseScreenCoords = { x: s.mouseX, y: s.mouseY };

      marblePlaceholder.hover(mouseWorldCoords);
      marblePlaceholder.update(mouseWorldCoords);
      marblePlaceholder.show(
        worldCoordsToScreenCoords(s, marblePlaceholder.x, marblePlaceholder.y),
      );
      powerPlaceholder.hover(mouseWorldCoords);
      powerPlaceholder.update(mouseWorldCoords);
      powerPlaceholder.show(
        worldCoordsToScreenCoords(s, powerPlaceholder.x, powerPlaceholder.y),
      );

      drawGameField(s);
      drawOpponentCursor(s, opponentCursorPositionHistory);

      // draw marbles
      /** @type {M.MarbleGameFrame} */
      let frame;
      if (frameIndex == -1) {
        frame = game.frames[game.frames.length - 1];
      } else {
        frame = game.frames[frameIndex];
      }
      drawMarbles(s, frame);
      if (frameIndex != -1) {
        frameIndex++;
      }
      if (frameIndex >= game.frames.length) {
        frameIndex = -1;
      }

      s.translate(0, 0, 600);
      drawPlayerScores(s, game);
      drawInventory(s);

      if (isMyTurn) {
        const player = game.players[userToken];
        const playerColor = s.color(`hsb(${player.hue},50%,100%)`);
        if (s.mouseIsPressed && s.mouseButton == s.LEFT) {
          // draw the placeholder ball with a line
          if (player.inventory.length > 0) {
            s.push();
            const mousePressedScreenCoords = worldCoordsToScreenCoords(
              s,
              mousePressedWorldCoords.x,
              mousePressedWorldCoords.y,
            );
            s.translate(mousePressedScreenCoords.x, mousePressedScreenCoords.y);
            s.stroke(playerColor);
            s.fill(255, 255, 255, 50);
            s.circle(
              0,
              0,
              game.players[userToken].inventory[selectedInventorySlot].radius *
                2,
            );
            s.textAlign(s.CENTER);
            s.pop();

            s.push();
            const diffX = mouseWorldCoords.x - mousePressedWorldCoords.x;
            const diffY = mouseWorldCoords.y - mousePressedWorldCoords.y;
            let powerMagnitude = Math.sqrt(diffX * diffX + diffY * diffY);
            const maxPower = 215;
            if (powerMagnitude > maxPower) {
              powerMagnitude = maxPower;
            }
            let powerPercentage = s.map(powerMagnitude, 0, maxPower, 0, 1);
            let from = s.color(`hsb(100,90%,100%)`);
            let to = s.color(`hsb(0,90%,100%)`);
            let color = s.lerpColor(from, to, powerPercentage);
            s.stroke(color);
            s.strokeWeight(3);
            if (powerMagnitude == maxPower) {
              s.strokeWeight(6);
            }
            s.strokeCap(s.ROUND);

            let x = mouseWorldCoords.x;
            let y = mouseWorldCoords.y;
            if (x < 0) {
              x = 0;
            }
            if (x > 600) {
              x = 600;
            }
            if (y < 0) {
              y = 0;
            }
            if (y > 480) {
              y = 480;
            }
            let velVec = s.createVector(
              x - mousePressedWorldCoords.x,
              y - mousePressedWorldCoords.y,
            );

            let velMag = velVec.mag();
            if (velMag > maxPower) {
              velMag = maxPower;
            }
            let velNormal = velVec.normalize();
            let visualPoint = velNormal.mult(velMag);

            s.line(
              mousePressedScreenCoords.x,
              mousePressedScreenCoords.y,
              mousePressedScreenCoords.x + visualPoint.x,
              mousePressedScreenCoords.y + visualPoint.y,
            );
            s.circle(
              mousePressedScreenCoords.x + visualPoint.x,
              mousePressedScreenCoords.y + visualPoint.y,
              10,
            );
            s.strokeWeight(1);
            drawDashedLine(
              s,
              mousePressedScreenCoords.x,
              mousePressedScreenCoords.y,
              mousePressedScreenCoords.x - visualPoint.x,
              mousePressedScreenCoords.y - visualPoint.y,
              10,
              10,
            );
            s.pop();
          }
        } else {
          // draw the placeholder ball
          if (player.inventory.length > 0) {
            s.push();
            s.stroke(playerColor);
            s.translate(Math.round(s.mouseX), Math.round(s.mouseY));
            s.fill(255, 255, 255, 50);
            s.circle(
              0,
              0,
              game.players[userToken].inventory[selectedInventorySlot].radius *
                2,
            );
            s.textAlign(s.CENTER);
            s.pop();
          }
        }
      }
    }
  };

  s.mouseMoved = function () {
    const worldCoords = screenCoordsToWorldCoords(s, s.mouseX, s.mouseY);

    const mouseXInput = window.document.getElementById("mouseX");
    mouseXInput.value = Math.floor(worldCoords.x);
    const mouseYInput = window.document.getElementById("mouseY");
    mouseYInput.value = Math.floor(worldCoords.y);

    cursorForm.dispatchEvent(new Event("sendit"));
  };

  s.mouseWheel = function (event) {
    if (event.delta > 0) {
      selectedInventorySlot++;
    }
    if (event.delta < 0) {
      selectedInventorySlot--;
    }
    const player = game.players[userToken];
    const inventoryLength = player.inventory.length;
    if (selectedInventorySlot < 0) {
      selectedInventorySlot = inventoryLength - 1;
    }
    if (selectedInventorySlot > inventoryLength - 1) {
      selectedInventorySlot = 0;
    }

    return false;
  };

  s.mouseDragged = function (event) {
    if (s.mouseButton == s.CENTER) {
      worldViewOffset.x += event.movementX;
      worldViewOffset.y += event.movementY;
    }
  };

  s.mousePressed = function () {
    if (s.mouseButton == s.LEFT) {
      mousePressedWorldCoords.x = Math.round(mouseWorldCoords.x);
      mousePressedWorldCoords.y = Math.round(mouseWorldCoords.y);
      marblePlaceholder.pressed(mouseWorldCoords);
      powerPlaceholder.pressed(mouseWorldCoords);
    }
  };

  s.mouseReleased = function () {
    /** @type {M.Action} */
    if (s.mouseButton == s.LEFT) {
      marblePlaceholder.released(mouseWorldCoords);
      powerPlaceholder.released(mouseWorldCoords);

      const action = {
        userToken: userToken,
        pos: {
          X: mousePressedWorldCoords.x,
          Y: mousePressedWorldCoords.y,
        },
        vel: {
          X: Math.round(mouseWorldCoords.x),
          Y: Math.round(mouseWorldCoords.y),
        },
        inventorySlot: selectedInventorySlot,
      };

      const actionInput = window.document.getElementById("action");
      actionInput.value = JSON.stringify(action);
      gameForm.dispatchEvent(new Event("sendit"));
    }
  };
}

function drawDashedLine(s, x1, y1, x2, y2, dashLength, gap) {
  let distance = s.dist(x1, y1, x2, y2);
  let dashCount = Math.floor(distance / (dashLength + gap));
  let dx = (x2 - x1) / distance;
  let dy = (y2 - y1) / distance;

  for (let i = 0; i < dashCount; i++) {
    let xStart = x1 + i * (dashLength + gap) * dx;
    let yStart = y1 + i * (dashLength + gap) * dy;
    let xEnd = xStart + dashLength * dx;
    let yEnd = yStart + dashLength * dy;

    s.line(xStart, yStart, xEnd, yEnd);
  }
}

/**
 * @param {p5} s
 * @param {number} x
 * @param {number} y
 */
function screenCoordsToWorldCoords(s, x, y) {
  return {
    x: x - worldViewOffset.x - s.width / 2 + game.config.width / 2,
    y: y - worldViewOffset.y - s.height / 2 + game.config.height / 2,
  };
}

/**
 * @param {p5} s
 * @param {number} x
 * @param {number} y
 */
function worldCoordsToScreenCoords(s, x, y) {
  return {
    x: x + worldViewOffset.x + s.width / 2 - game.config.width / 2,
    y: y + worldViewOffset.y + s.height / 2 - game.config.height / 2,
  };
}

addEventListener("htmx:wsBeforeMessage", (event) => {
  const { elt, message } = event.detail;

  if (elt == cursorForm) {
    // got updates on opponent's cursor pos
    const json = JSON.parse(message);
    opponentCursorPositionHistory.push({
      userToken: json.userToken,
      x: parseInt(json.mouseX),
      y: parseInt(json.mouseY),
    });
    if (opponentCursorPositionHistory.length > 10) {
      opponentCursorPositionHistory.shift();
    }
  }
  if (elt == gameForm) {
    // got updates on game state

    // reset inventorySlot
    selectedInventorySlot = 0;

    /** @type {M.MarbleGame} */
    const json = JSON.parse(message);
    console.log(json);
    game = json;

    // check if it's my turn
    isMyTurn = json.turnOrder[json.activePlayerIndex].userToken == userToken;

    // now playback all that jazz
    frameIndex = 0;
  }
});

/**
 * @param {p5} s
 * @param {M.MarbleGame} game
 */
function drawPlayerScores(s, game) {
  let offset = 0;
  for (const [playerUserToken, player] of Object.entries(game.players)) {
    s.push();
    const color = s.color(`hsb(${player.hue},50%,100%)`);
    s.fill(color);
    s.stroke("black");
    if (playerUserToken == userToken) {
      s.textStyle(s.BOLD);
    }
    s.strokeWeight(1);
    s.textAlign(s.CENTER);
    s.translate(s.width / 2, 30);
    s.text(
      `${game.turnOrder[game.activePlayerIndex].userToken == playerUserToken ? "> " : ""}${playerUserToken == userToken ? "(you)" : player.userToken.slice(0, 4)}: ${player.score}`,
      0,
      offset,
    );
    s.pop();
    offset -= 12;
  }
}

/**
 * @param {p5} s
 * @param {CursorPosition[]} opponentCursorPositionHistory
 */
function drawOpponentCursor(s, opponentCursorPositionHistory) {
  for (let i = 0; i < opponentCursorPositionHistory.length; i++) {
    if (i < opponentCursorPositionHistory.length - 1) {
      s.push();
      s.stroke(100);
      const screenCoords = worldCoordsToScreenCoords(
        s,
        opponentCursorPositionHistory[i].x,
        opponentCursorPositionHistory[i].y,
      );
      const screenCoords1 = worldCoordsToScreenCoords(
        s,
        opponentCursorPositionHistory[i + 1].x,
        opponentCursorPositionHistory[i + 1].y,
      );
      s.line(screenCoords.x, screenCoords.y, screenCoords1.x, screenCoords1.y);
      s.pop();
    }
  }
  if (opponentCursorPositionHistory.length > 0) {
    const finalPosition =
      opponentCursorPositionHistory[opponentCursorPositionHistory.length - 1];
    s.push();
    const screenCoords = worldCoordsToScreenCoords(
      s,
      finalPosition.x,
      finalPosition.y,
    );
    s.translate(screenCoords.x, screenCoords.y);
    s.image(cursorImg, -5, -5, 30, 30);
    s.stroke(0, 100);
    s.fill(255, 100);
    s.text(finalPosition.userToken.slice(0, 4), 22, 20);
    s.pop();
  }
}

/**
 * @param {p5} s
 * @param {M.MarbleGameFrame} frame
 */
function drawMarbles(s, frame) {
  let shouldPlayCollisionSound = false;
  for (let i = 0; i < frame.marbles.length; i++) {
    const marble = frame.marbles[i];
    s.shader(shaderProgram);
    shaderProgram.setUniform("uTexture", marbleTextureImg);
    shaderProgram.setUniform("uQuaternion", [
      marble.rot[0],
      marble.rot[1],
      marble.rot[2],
      marble.rot[3],
    ]);
    if (marble.score == 0) {
      shaderProgram.setUniform("uOpacity", 0.5);
    } else {
      shaderProgram.setUniform("uOpacity", 1.0);
    }
    shaderProgram.setUniform("uLightDir", [
      -marble.pos.X,
      -s.height,
      s.width * 3.0,
    ]);

    const marbleScreenCoords = worldCoordsToScreenCoords(
      s,
      marble.pos.X,
      marble.pos.Y,
    );
    const marbleNormalizedX = s.map(marbleScreenCoords.x, 0, s.width, -1, 1);
    const marbleNormalizedY = s.map(marbleScreenCoords.y, 0, s.height, -1, 1);
    const width = s.map(marble.type.radius * 2, 0, s.width, 0, 2);
    const height = s.map(marble.type.radius * 2, 0, s.height, 0, 2);
    const bl = {
      x: marbleNormalizedX - width / 2,
      y: marbleNormalizedY + height / 2,
    };
    const br = {
      x: marbleNormalizedX + width / 2,
      y: marbleNormalizedY + height / 2,
    };
    const tr = {
      x: marbleNormalizedX + width / 2,
      y: marbleNormalizedY - height / 2,
    };
    const tl = {
      x: marbleNormalizedX - width / 2,
      y: marbleNormalizedY - height / 2,
    };
    s.beginShape();
    s.vertex(bl.x, -bl.y, 0, 0, 1);
    s.vertex(br.x, -br.y, 0, 1, 1);
    s.vertex(tr.x, -tr.y, 0, 1, 0);
    s.vertex(tl.x, -tl.y, 0, 0, 0);
    s.endShape(s.CLOSE);
    s.resetShader();

    s.push();
    const playerColor = s.color(`hsb(${marble.owner.hue},50%,100%)`);
    s.stroke(playerColor);
    s.translate(marbleScreenCoords.x, marbleScreenCoords.y, 500);
    s.noFill();
    if (marble.collided) {
      s.circle(0, 0, marble.type.radius * 2.1);
    } else {
      s.circle(0, 0, marble.type.radius * 2);
    }
    s.textAlign(s.CENTER);
    if (marble.score != 0) {
      s.fill(0);
      s.text(marble.score, 1, 6);
      s.fill(255);
      s.text(marble.score, 0, 5);
    }
    s.pop();

    // play collision sound
    if (marble.collided) {
      shouldPlayCollisionSound = true;
    }
  }
  if (shouldPlayCollisionSound) {
    if (canPlayAudio) {
      hitSound.play();
    }
  }
}

/**
 * @param {p5} s
 */
function drawGameField(s) {
  s.push();
  const center = worldCoordsToScreenCoords(
    s,
    game.config.width / 2,
    game.config.height / 2,
  );
  s.translate(center.x, center.y);
  s.noFill();
  s.stroke(100);
  s.circle(0, 0, game.config.scoringZoneRadius * 2);
  s.circle(0, 0, game.config.bullseyeZoneRadius * 2);
  s.rectMode(s.CENTER);
  s.rect(0, 0, game.config.width, game.config.height);
  s.pop();
}

/**
 * @param {p5} s
 */
function drawInventory(s) {
  // const you = game.players.find((p) => p.userToken == userToken);
  const you = game.players[userToken];
  let offset = 0;
  for (let i = 0; i < you.inventory.length; i++) {
    const marbleType = you.inventory[i];
    s.push();
    s.translate(10, s.height - 20 - offset);
    if (selectedInventorySlot == i) {
      s.fill(255);
    } else {
      s.fill(100);
    }
    s.rect(0, 0, 8, 8);
    s.pop();
    offset += 20;
  }
}

new window.p5(mySketch);
