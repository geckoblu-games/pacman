# Technical Specifications: Pac-Man Attract Mode (Introductory Screen)

This document details the spatial layout, timeline, and state machine of the introductory screen (**Attract Mode**) from the original arcade version of *Pac-Man* (Namco / Midway, 1980), designed to serve as a direct reference for implementation.

**Video reference:** [Pac-Man Attract Mode](https://www.youtube.com/watch?v=yqFPcaY-qmQ)

---

## 1. Base Screen & Grid Parameters

* **Native Resolution:** `224 x 288` pixels (vertical 3:4 aspect ratio).
* **Coordinate System:** The origin `(0, 0)` is located at the top-left corner.
* **Logical Grid (Tile Map):** `28 x 36` tiles, where each tile measures `8 x 8` pixels.
* **Refresh Rate (Frame Rate):** `60.606061 Hz` (modern clones typically use `60 FPS`, where `60 frames = 1.0 second`).
* **Sprite Dimensions:** All character sprites measure `16 x 16` pixels (centered relative to their logical coordinates).
* **Audio:** The entire Attract Mode cycle is completely silent (no sound effects or background music).

---

## 2. Static Layout & Coordinates (28x36 Grid)

At the start of the cycle, the screen is completely cleared (black background `#000000`) and the top and bottom UI elements are rendered, remaining static throughout the presentation.

### 2.1 Top Header (HUD)

| Element | Text | Color | Tile Position `(Col, Row)` | Pixel Position `(X, Y)` |
| :--- | :--- | :--- | :--- | :--- |
| **1UP Label** | `1UP` | White (`#DEDEDE`) | `(3, 0)` | `(24, 0)` |
| **High Score Label** | `HIGH SCORE` | White (`#DEDEDE`) | `(9, 0)` | `(72, 0)` |
| **2UP Label** | `2UP` | White (`#DEDEDE`) | `(22, 0)` | `(176, 0)` |
| **1UP Score** | `00` (or last score) | White (`#DEDEDE`) | `(5, 1)` (right-aligned) | `(40, 8)` |
| **High Score Value** | Current record | White (`#DEDEDE`) | `(16, 1)` (right-aligned) | `(128, 8)` |
| **Table Header** | `CHARACTER / NICKNAME` | White (`#DEDEDE`) | `(7, 5)` | `(56, 40)` |

### 2.2 Character Cast Table
Each presentation row occupies a vertical space of 3 tiles (`24 pixels`). The ghost sprite is rendered facing right.

| Ghost | Color | Sprite Position `(X, Y)` | Character Text | Character Pos `(Col, Row)` | Nickname Text | Nickname Pos `(Col, Row)` |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Blinky** | Red (`#FF0000`) | `(36, 56)` | `- SHADOW` | `(7, 7)` | `"BLINKY"` | `(18, 7)` |
| **Pinky** | Pink (`#FFB8FF`) | `(36, 80)` | `- SPEEDY` | `(7, 10)` | `"PINKY"` | `(18, 10)` |
| **Inky** | Cyan (`#00FFFF`) | `(36, 104)` | `- BASHFUL` | `(7, 13)` | `"INKY"` | `(18, 13)` |
| **Clyde** | Orange (`#FFB852`) | `(36, 128)` | `- POKEY` | `(7, 16)` | `"CLYDE"` | `(18, 16)` |

### 2.3 Point Legend & Footer

| Element | Graphic / Text | Color | Tile Position `(Col, Row)` | Pixel Position `(X, Y)` |
| :--- | :--- | :--- | :--- | :--- |
| **Dot Icon** | Dot sprite (`2x2` px) | Peach pink (`#FFB897`) | `(10, 24)` | `(80, 192)` |
| **Dot Text** | `10 PTS` (with small `PTS`) | White (`#DEDEDE`) | `(12, 24)` | `(96, 192)` |
| **Energizer Icon** | Power Pellet sprite (`8x8` px) | Peach pink (`#FFB897`) | `(10, 26)` | `(80, 208)` |
| **Energizer Text** | `50 PTS` (with small `PTS`) | White (`#DEDEDE`) | `(12, 26)` | `(96, 208)` |
| **Copyright** | `© 1980 MIDWAY MFG.CO.`* | Pink (`#FFB8FF`) | `(4, 31)` | `(32, 248)` |
| **Credits** | `CREDIT  0` | White (`#DEDEDE`) | `(2, 35)` | `(16, 280)` |

*\*Note: In the original Japanese Namco release, the text reads `© 1980 NAMCO LTD.`.*

---

## 3. Sequential Timeline (Frame & Second Timings)

The black-screen presentation phase lasts a total of **approximately 19.5 seconds (1,170 frames at 60 FPS)** before transitioning to the demo gameplay.

### Phase A: Character Reveal (`0.0s` – `8.5s`)
1. **Frame `0` (`0.0s`):** Screen clear. Top HUD, `CHARACTER / NICKNAME`, Copyright, and `CREDIT 0` appear.
2. **Frame `60` (`1.0s`):** **Blinky**'s sprite appears at `(36, 56)`.
3. **Frame `120` (`2.0s`):** The text `- SHADOW` and `"BLINKY"` appear.
4. **Frame `150` (`2.5s`):** **Pinky**'s sprite appears at `(36, 80)`.
5. **Frame `210` (`3.5s`):** The text `- SPEEDY` and `"PINKY"` appear.
6. **Frame `240` (`4.0s`):** **Inky**'s sprite appears at `(36, 104)`.
7. **Frame `300` (`5.0s`):** The text `- BASHFUL` and `"INKY"` appear.
8. **Frame `330` (`5.5s`):** **Clyde**'s sprite appears at `(36, 128)`.
9. **Frame `390` (`6.5s`):** The text `- POKEY` and `"CLYDE"` appear.
10. **Frame `450` (`7.5s`):** The point legend (`10 PTS` and `50 PTS`) and a **decoy Power Pellet** placed on the left side of the virtual corridor at pixel coordinates `(24, 160)` (Tile `Col 3, Row 20`) appear simultaneously.

### Phase B: The Chase (`8.5s` – `12.5s`)
1. **Frame `510` (`8.5s`):** From this moment on, the decoy Power Pellet at `(24, 160)` and the legend Power Pellet at `(80, 208)` begin flashing on a **20-frame cycle** (10 frames visible, 10 frames invisible).
2. **Stage Entrance (Frame `510`):** The actors enter from off-screen right (`Y = 160`, aligned with row 20) moving left (`-X`):
   * **Pac-Man** starts at `X = 236` (speed: ~`1.35 px/frame`).
   * **Blinky** starts at `X = 258` (speed: ~`1.42 px/frame`, slightly faster than Pac-Man to close the gap).
   * **Pinky** starts at `X = 276` (speed: ~`1.42 px/frame`).
   * **Inky** starts at `X = 294` (speed: ~`1.42 px/frame`).
   * **Clyde** starts at `X = 312` (speed: ~`1.42 px/frame`).
3. **Frame `666` (`11.1s`):** Pac-Man reaches `X = 24` and eats the decoy Power Pellet.

### Phase C: Reversal & Capture (`11.1s` – `19.5s`)
1. **Frame `666` (`11.1s`):**
   * The decoy Power Pellet disappears.
   * All 4 ghosts turn **Blue (Frightened)**, immediately reverse direction to the right (`+X`), and reduce their speed to approximately `0.65 px/frame`.
   * Pac-Man reverses direction to the right (`+X`), maintaining a speed of approximately `1.35 px/frame`.
2. **Ghost Capture Sequence:** When Pac-Man's center (`X`) reaches the center of each ghost, the game triggers a **45-frame pause (`0.75s`)** during which both Pac-Man and the ghost disappear, leaving only the score sprite visible:
   * **1st Ghost (Blinky):** Captured around frame `720` (`12.0s`). `200` appears for `0.75s`. Blinky disappears permanently (does not turn into eyes).
   * **2nd Ghost (Pinky):** Captured around frame `810` (`13.5s`). `400` appears for `0.75s`. Pinky disappears.
   * **3rd Ghost (Inky):** Captured around frame `900` (`15.0s`). `800` appears for `0.75s`. Inky disappears.
   * **4th Ghost (Clyde):** Captured around frame `990` (`16.5s`). `1600` appears for `0.75s`. Clyde disappears.
3. **Frame `1035` (`17.25s`):** The `1600` score text disappears and Pac-Man completes his exit or holds position awaiting the transition.
4. **Frame `1170` (`19.5s`):** Full screen clear and transition to **Phase D (Demo Play)**.

---

## 4. Phase D: Demo Play (Simulated Gameplay)

At the end of the introductory sequence, the game loads the standard **Level 1** maze with the following Attract Mode-specific modifications:

* **Game Over Banner:** Instead of the `READY!` text below the ghost house (`Col 9, Row 20`), the red text **`GAME OVER`** is displayed.
* **Lives:** The life counter at the bottom left displays no Pac-Man icons (0 lives shown).
* **Pre-recorded Inputs:** Pac-Man moves by executing a hardcoded direction table stored in the ROM (a fixed sequence of turns based on a timer/position).
* **Ghost Behavior:** The four ghosts use their standard Level 1 AI (*Scatter* and *Chase* cycles, including exiting the ghost house based on the dot counter).
* **End of Cycle:** When a ghost catches Pac-Man during the demo, the standard death animation plays out. Once the animation finishes, instead of restarting the level, the state machine immediately jumps back to **Frame 0 of Phase A**, repeating the entire loop indefinitely.

---

## 5. Recommended Code Architecture (State Machine)

To implement this screen cleanly and modularly, isolate the Attract Mode inside a dedicated state (`GameState.ATTRACT_MODE`) driven by a frame counter (`attractTimer++` on every fixed 60Hz tick):

```text
ATTRACT_MODE
 ├── SUBSTATE_CAST_REVEAL   (Frame 0 -> 509: sequential timer for text and sprites)
 ├── SUBSTATE_CHASE_LEFT    (Frame 510 -> 665: update X positions moving left)
 ├── SUBSTATE_CHASE_RIGHT   (Frame 666 -> 1169: move right + 45-frame pause timer per capture)
 └── SUBSTATE_DEMO_PLAY     (Load Level 1 with isDemo = true flag; on Game Over reset attractTimer = 0)
```

### Interrupt Rule (Coin Insert)
On any frame of any substate, if an `INSERT_COIN` event (`credits > 0`) is detected:
1. Immediately halt the Attract Mode update loop.
2. Play the coin insert sound effect.
3. Transition the global game state to `GameState.START_SCREEN` (the orange `PUSH START BUTTON` screen with `1 PLAYER ONLY` / `1 OR 2 PLAYERS` selection).

---

## 6. Project Implementation Notes

The preceding sections describe the arcade reference. This project intentionally differs in the following ways:

* **Simulation and scale:** The game advances at 60 ticks per second. Arcade coordinates and speeds in this document use 8-pixel tiles; the renderer scales them to its 16-pixel logical tiles and a `448 x 576` logical screen.
* **End of the attract cycle:** At tick `1170`, the presentation restarts from Frame 0 instead of entering Phase D. The ROM-recorded demo play and its death/restart loop are not implemented.
* **Starting a match:** `ENTER` replaces coin insert and starts a single-player match directly with the `READY!` intro. There is no credit counter, `PUSH START BUTTON` screen, or player-count selection. The `2UP` label is decorative; a second player is not available.
* **Audio:** The attract presentation is silent, as in the arcade. By project decision, `ENTER` starts the match without a coin-insert effect; the opening `READY!` jingle begins at once.
* **Timeline:** The configured movement speeds determine the actual event ticks: Pac-Man eats the decoy at tick `670`, and the four captures occur at ticks `686`, `757`, `827`, and `898`. These differ from the approximate frame numbers in the arcade timeline above.
* **Cast-table typography:** The displayed character descriptions include a space after the hyphen for legibility. Their icon alignment follows the companion [PNG reference](./Pac-Man%20Attract%20Mode%20(Introductory%20Screen).png).

