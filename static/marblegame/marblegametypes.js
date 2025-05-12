/**
 * Represents a marble game.
 * @typedef {Object} MarbleGame
 * @property {Object.<string, Player>} players - A map of player IDs to Player objects.
 * @property {MarbleGameFrame[]} frames - The history of game frames.
 * @property {MarbleGameConfig} config - The configuration settings of the game.
 */

/**
 * Configuration settings for the marble game.
 * @typedef {Object} MarbleGameConfig
 * @property {number} playerLimit - Maximum number of players allowed.
 * @property {number} scoringZoneRadius - The radius of the scoring zone.
 * @property {number} scoringZoneMaxScore - The maximum score possible in the scoring zone.
 * @property {number} scoringZoneMinScore - The minimum score possible in the scoring zone.
 * @property {number} bullseyeZoneRadius - The radius of the bullseye zone.
 * @property {number} bullseyeZoneScore - The score for hitting the bullseye zone.
 * @property {number} width
 * @property {number} height
 */

/**
 * Represents a single frame in the game.
 * @typedef {Object} MarbleGameFrame
 * @property {Marble[]} marbles - The current state of marbles in the game.
 */

/**
 * Represents a player in the game.
 * @typedef {Object} Player
 * @property {string} userToken - A unique token for the player (not serialized in JSON).
 * @property {string} displayName - The player's display name.
 * @property {number} score - The player's current score.
 * @property {number} hue - The player's color hue.
 * @property {boolean} isTheirTurn - Whether it is currently the player's turn.
 * @property {MarbleType[]} inventory - The player's inventory of marble types.
 */

/**
 * Represents an action taken by a player.
 * @typedef {Object} Action
 * @property {number} inventorySlot - The inventory slot of the player being used.
 * @property {Vector2} pos - The position of the marble.
 * @property {Vector2} vel - The velocity of the marble.
 * @property {string} userToken - The token of the user performing the action.
 */

/**
 * Represents a marble in the game.
 * @typedef {Object} Marble
 * @property {Vector2} pos - The position of the marble.
 * @property {Vector2} vel - The velocity of the marble.
 * @property {number[]} rot - The rotation of the marble.
 * @property {number} score - The score assigned to the marble.
 * @property {MarbleType} type - The type of the marble.
 * @property {boolean} collided - Whether the marble has collided.
 * @property {string} highlightColor - The color used to highlight the marble.
 * @property {Player|null} owner - The player who owns this marble (nullable).
 */

/**
 * Represents a type of marble.
 * @typedef {Object} MarbleType
 * @property {string} name - The name of the marble type.
 * @property {string} description - A description of the marble type.
 * @property {number} radius - The radius of the marble.
 * @property {number} mass - The mass of the marble.
 */

/**
 * Represents a 2D vector.
 * @typedef {Object} Vector2
 * @property {number} X - The X coordinate.
 * @property {number} Y - The Y coordinate.
 */

module.exports = {};
