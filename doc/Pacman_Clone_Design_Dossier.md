# Pac-Man Clone — Functional Analysis & Game Design Document
### Technical Reference for a Go + Ebiten Implementation

**Primary source:** *The Pac-Man Dossier* , chapters 2–4 and Appendix A, as provided in `The Pac-Man Dossier.md`.

**Purpose:** This document translates the dossier's behavioral analysis of the original arcade game into a structured set of functional requirements and architectural guidelines for building a faithful Pac-Man clone in Go using the Ebiten 2D engine.

It intentionally contains **no code or pseudocode** — it is a design and requirements reference to be used before and during implementation.

**Convention used throughout this document:**
- **[FACT]** — stated explicitly in the dossier, treated as ground truth.
- **[ASSUMPTION]** — not fully specified by the dossier; a reasonable design decision is proposed and clearly separated from the facts.
- **[AMBIGUITY]** — a point where the dossier's wording is open to interpretation, incomplete, or internally inconsistent; the implication for the implementation is stated.
- **[DECISION]** — an owner-approved gameplay choice made where source material is incomplete or ambiguous.

---

## 1. Game Rules, Objectives, and Win/Lose Conditions

### 1.1 Core Loop
- **[FACT]** The player steers Pac-Man through a maze using four-way directional input (up/down/left/right). The objective each round is to eat all the dots in the maze while avoiding four ghosts that try to capture him.
- **[FACT]** When the maze is fully cleared of dots, the board resets and a new round (level) begins, with increased difficulty.
- **[FACT]** If a ghost captures Pac-Man (occupies the same tile — see §1.4 and §7.3), one life is lost; ghosts return to the ghost house, and Pac-Man resets to his starting position before play resumes.
- **[FACT]** The game ends when Pac-Man is captured while having zero extra lives remaining ("Game Over").

### 1.2 Dot Count and Maze Composition
- **[FACT]** The reference maze contains 244 dots: 240 regular dots (10 points each) and 4 **energizers** (50 points each, located near the maze corners), yielding 2,600 points for a fully cleared maze.

### 1.3 Score Sources
Three sources of points, matching the dossier:
1. Regular dots — 10 pts each.
2. Energizers — 50 pts each.
3. Ghosts eaten while frightened — 200 / 400 / 800 / 1,600 pts, doubling per ghost eaten from the *same* energizer activation (resets to 200 on the next energizer). **[FACT]**
4. Bonus fruit — 100 to 5,000 pts depending on level (see §5.3). **[FACT]**

- **[DECISION]** Reset the ghost score sequence when frightened time expires or another energizer is eaten. The 12,000 points for eating all four ghosts with each of the four energizers are the sum of those sixteen awards, not an additional secret bonus, in accordance with the source dossier.
- **[DECISION]** Capturing a frightened ghost freezes gameplay for 60 ticks. Pac-Man and the captured ghost are hidden, and the awarded points appear at the ghost's collision position; other ghosts remain visible and stationary. Frightened, mode, fruit, and actor timers do not advance. On resumption Pac-Man and the returning eyes appear at their collision positions. The capture plays the ghost-eaten effect; the background loop stops during the freeze.

### 1.4 Lose Condition (Collision)
- **[FACT]** Collision is defined purely by **tile occupancy**: whenever Pac-Man and a ghost share the same tile (regardless of who moved into whose tile), a life is lost — *except* for the rare "pass-through" edge case (§7.3), which the dossier documents as an authentic artifact of the original's collision detection, not a bug to "fix" for faithfulness purposes.
- **[AMBIGUITY]** The dossier does not state whether a frightened or already-reversing ghost that shares a tile with Pac-Man as it is being eaten counts differently. **[ASSUMPTION]** Treat ghost state at the moment of tile-sharing as authoritative: *frightened* → ghost is eaten (see §4), *chase/scatter* → Pac-Man loses a life. This must be evaluated with the ghost's **current** mode on the frame of collision.
- **[DECISION]** A fatal collision freezes the entire scene for 60 ticks, then hides ghosts and fruit and plays a 120-tick Pac-Man death animation including a final spark. The maze remains without actors for 60 more ticks; only then is a life deducted. Remaining lives lead to the normal READY intro, while the last life leads to a centered red `GAME OVER`. Collectibles already eaten remain consumed.
- **[DECISION]** The collision freeze is silent; the death sound starts with the animation and lasts its 2 seconds. The animation is procedurally generated and is not an exact reproduction of the arcade sprites.

### 1.5 Victory/Progression Condition
- **[FACT]** Clearing all dots in the maze (240 + 4) ends the round and advances the level counter; there is no final "win" screen described in the dossier — the game is designed to continue indefinitely with escalating/stabilizing difficulty (see §5).
- **[ASSUMPTION]** For a clone, decide explicitly whether to cap levels (e.g., loop content after level 21+, matching the "steady state" difficulty table) or implement procedurally generated/varied mazes beyond a certain level. The dossier's data stabilizes at level 21+, implying the original design intends indefinite play at a fixed difficulty ceiling — this is a reasonable stopping point for parity.

### 1.6 Level-Clear and New-Level Intro
- **[DECISION]** When the last collectible is consumed, the completed maze flashes blue and white for exactly 2 seconds before the next level begins. The colors alternate for a total of four equal flashes during this interval.
- **[DECISION]** Pac-Man stops immediately where he is when the level-clear flash begins, and all four ghosts disappear from the screen immediately. The maze flash is the only visible level-clear effect specified here.
- **[DECISION]** At the start of the next level, show a centered yellow `READY!` message for approximately 2 seconds. Pac-Man and all four ghosts are visible, stationary, and placed at their starting positions during this intro.
- **[DECISION]** Play the start jingle only during a match's opening `READY!` intro, as the arcade does; intros after a lost life or a cleared level are silent and keep the ~2-second duration. The opening intro lasts 4.4 seconds (264 ticks), covering the 4.34-second jingle, so gameplay controls stay locked and all characters stationary until `READY!` has disappeared and the jingle has finished; then controls unlock and the characters move.
- **[DECISION]** The flash is split into eight equal 15-tick segments, starting white as soon as the last collectible is eaten and ending blue, so the maze is blue again when the next level's intro appears.
- **[DECISION]** The same `READY!` intro (centered yellow message, controls locked, automatic start) is shown at the start of a new match and after a life is lost, not only between levels; only its duration and jingle differ as described above. No key press is required to begin play.
- **Timing note:** The simulation uses a 60 Hz logical tick (§9.2), so the exact 2-second level-clear flash corresponds to 120 ticks. The later READY duration is approximate, not an exact tick count; the opening one is fixed by the jingle length.

### 1.7 Persistent High Score
- **[DECISION]** Preserve the high score across application runs as an intentional desktop convenience, not an arcade-fidelity requirement inferred from the source dossier. Starting a new match resets its score but retains the record.
- **[DECISION]** On Linux, use `$XDG_DATA_HOME/geckoblu/pacman/highscore` when `XDG_DATA_HOME` is absolute; otherwise use `~/.local/share/geckoblu/pacman/highscore`.
- **[DECISION]** Use `%LOCALAPPDATA%\geckoblu\pacman\highscore` on Windows and `~/Library/Application Support/geckoblu/pacman/highscore` on macOS. On iOS, use Application Support inside the application's sandbox. Android hosts must supply a record file inside the app-private persistent files directory; do not infer package-specific paths or use shared storage.
- **[DECISION]** Mobile support at this stage is limited to the persistence module. A native host must supply Android's private path and wire mobile loading, checkpoints, and lifecycle saves; mobile packaging and lifecycle integration are not implemented.
- **[DECISION]** Save a new record immediately upon level completion (before the maze flash), at game over, and on normal window closure. Do not write during ordinary scoring or rewrite an unchanged record.
- **[DECISION]** Report persistence errors without preventing play. Keep an unsaved record in memory and retry at the next checkpoint. Abrupt termination can lose points earned since the last successful checkpoint.

### 1.8 Pause
- **[DECISION]** Pressing `P` pauses the match as a desktop convenience, not an arcade-fidelity feature. Pressing `P` again resumes. While paused, the whole simulation is frozen (phase timers, bonus timers, movement, and gameplay input) and resumes exactly where it stopped, so pausing does not alter any tick-counted behavior.
- **[DECISION]** Pausing is possible during the READY! intro, active play, the ghost-capture pause, the death sequence, and the level-clear flash. It is rejected when no match exists or after game over, and a new match always starts unpaused.
- **[DECISION]** While paused, a centered yellow `PAUSED` message replaces the `READY!` message position on the corridor below the ghost house; the rest of the screen is drawn unchanged.
- **[DECISION]** Pausing also pauses every playing sound, including the opening jingle and the background loop, and resuming continues them in place.

### 1.10 Sound
- **[FACT]** The arcade plays a start jingle, an alternating waka for each dot, a background siren whose pitch rises as dots are eaten, a distinct loop while ghosts are frightened, another while eaten ghosts' eyes return, and one-shot effects for eating a fruit or a ghost, for Pac-Man's death, for coin insertion and for the extra life.
- **[DECISION]** Use the recordings published by the Warhorst/pacman project (credited in the README). They provide a single waka bite, a single siren, the frightened and retreating loops, and the start, fruit, ghost and death effects; no coin-insert or extra-life sound is available.
- **[DECISION]** Generate the siren's rising pitch from the single recording by resampling it at speeds 1.0, 1.05, 1.1 and 1.15, selecting one level per eaten quarter of the level's 244 collectibles. This is an approximation of the arcade's distinct siren recordings, chosen because no separate recordings are available.
- **[DECISION]** The background loop is a function of the current state: the retreating loop while any eyes return, otherwise the frightened loop while ghosts are frightened, otherwise the siren; it plays only during unpaused active play. The attract screen, the `READY!` intros, the capture freeze, the death sequence, the level-clear flash and `GAME OVER` have no loop.
- **[DECISION]** Play one waka per eaten coin or energizer. Because only one recording of the bite exists and dots are eaten faster than it lasts, a bite requested while the previous one is playing is queued and played as soon as it ends; further requests during that time are merged into the queued one, so bites never overlap.
- **[DECISION]** Do not play the record-beaten jingle shipped with the sounds; it is not an arcade behavior. Score-awarded extra lives are not implemented (§5.6), so no extra-life sound is needed.
- **[DECISION]** `M` toggles mute as a desktop convenience; it silences every sound without stopping or desynchronizing playback.

### 1.9 Attract Screen and Match Start
- **[FACT]** The arcade's attract mode is specified in `Pac-Man Attract Mode (Introductory Screen).md`: a silent cast-reveal timeline, a corridor chase with a decoy energizer and four captures worth 200/400/800/1600, a demo play phase, and a coin-insert interrupt leading to a `PUSH START BUTTON` screen with a player-count choice.
- **[DECISION]** The application launches on the attract screen and reproduces its layout and timeline (phases A–C, 1,170 ticks) using the specification's 8-pixel-tile coordinates scaled to the 16-pixel grid. Sprite positions are read as top-left corners of 16-pixel sprites and converted to centers; the chase and the decoy share the center of row 20. Capture ticks follow from the specified speeds and the 45-tick pauses rather than the specification's approximate frame numbers. The copyright notice and the corridor decoy energizer appear when the chase starts, not during the cast reveal or legend-only interval.
- **[DECISION]** Demo play (phase D) is not implemented: at tick 1,170 the presentation restarts from its first tick and loops until a match starts.
- **[DECISION]** `Enter` replaces the coin insert. There is no credit counter, no `PUSH START BUTTON` screen and no two-player option: pressing `Enter` on the attract screen or after `GAME OVER` immediately starts a single-player match with the `READY!` intro. By project decision, this key press has no coin-insert effect; the start jingle begins at once.
- **[DECISION]** For arcade fidelity the attract screen shows the `2UP` label and a static `CREDIT  0` line, and the `1UP` field shows the previous match's final score (`00` at launch). Scores use the arcade's two-digit minimum (`00`) on every screen. The `2UP` label is also shown during play, as in the arcade; no second-player score is ever displayed.
- **[DECISION]** After the last life, `GAME OVER` remains visible for 300 ticks (five seconds, an assumption rather than a measured arcade value) and then returns to the attract screen; `Enter` during `GAME OVER` starts a new match at once.

---

## 2. Core Entities

| Entity                                  | Key Attributes (from dossier)                                                                                                       | Design Notes                                                                                                  |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| **Pac-Man (Player)**                    | Position (pixel-precise), direction, speed (% of max, context-dependent), lives, score                                              | Only actor capable of "cornering" (pre-/post-turns); immune to tunnel slow-down.                              |
| **Ghosts** (Blinky, Pinky, Inky, Clyde) | Position, direction, mode (chase/scatter/frightened), per-ghost dot counter & dot limit, unique targeting function, speed modifiers | Share one pathfinding algorithm; differ only in **target-tile calculation** (§4) and release conditions (§6). |
| **Regular dot**                         | Position, eaten/not-eaten, value = 10                                                                                               | Also drives the house-release counters and causes Pac-Man's 1-frame "eating pause."                           |
| **Energizer (power pellet)**            | Position (corner areas), eaten/not-eaten, value = 50                                                                                | Triggers frightened mode globally; causes a 3-frame eating pause for Pac-Man.                                 |
| **Bonus fruit**                         | Position (just below the ghost house), value (level-dependent, 100–5,000), lifespan (9–10 s, randomized), appears twice per round   | Not required to clear the level; purely a scoring opportunity.                                                |
| **Walls / maze structure**              | Static grid of "legal space" vs. "dead space" tiles                                                                                 | Defines navigable topology; also defines special zones (one-way "red zones", tunnel "pink zones" — §7).       |
| **Tunnel (side passage)**               | A pair of connected tiles at opposite horizontal maze edges                                                                         | Lets all actors wrap continuously; imposes a ghost-only speed penalty while inside.                           |
| **Ghost house ("monster pen")**         | A sealed region with one door tile; holds ghosts between release and recapture                                                      | Governs spawn/respawn logic and timing (§6).                                                                  |
| **Level**                               | Index (1, 2, 3, …), associated speed table, fright duration, fruit symbol/value, Elroy thresholds                                   | The master driver of nearly all tunable parameters (§5).                                                      |

---

## 3. Fundamental Mechanics and Their Interactions

### 3.1 Movement Granularity: Tiles vs. Pixels
- **[FACT]** The maze is a grid of 8×8-pixel tiles (28×36 tiles for the original 224×288 screen). Each tile is either **legal space** (navigable) or **dead space** (not navigable). Dots sit exactly at tile centers, one tile apart.
- **[FACT]** An actor's *occupied tile* is determined solely by where its **center point** currently lies, even though its sprite visually overlaps neighboring tiles. All ghost AI (pathfinding, target evaluation, decision-making) reasons strictly in terms of **tiles**, not raw pixels.
- **Design implication:** the implementation needs two coordinate systems that stay in sync — a continuous sub-pixel position for rendering and movement, and a discretized tile coordinate for logic/decisions/collisions. This separation is the single most important architectural fact to carry into the Go implementation (see §9).

### 3.2 Decision Points ("Looking Ahead")
- **[FACT]** A ghost does not decide "live" when it reaches an intersection. Instead, the moment it *enters* a tile, it immediately looks one tile further ahead (along its current direction) to determine the next direction it will commit to once it physically reaches that tile. Decisions are therefore always made **one tile in advance**.
- **[FACT]** At an intersection, a ghost evaluates all exits except the one that would reverse its current direction (reversal is normally forbidden — see §3.3). Among the remaining open exits, it computes, for each candidate, the tile one step beyond the intersection, and picks the direction whose resulting tile has the smallest Euclidean distance to its current **target tile**.
- **[FACT] Tie-break rule:** when two or more candidate directions yield equal distance to the target, the ghost prefers, in order: **Up > Left > Down > Right**.
- **[FACT]** This exact mechanism (target tile + lookahead + distance comparison + tie-break) is **shared by all four ghosts**, in both chase and scatter mode. The only thing that differs between ghosts, and between chase/scatter, is **where the target tile is located** (§4).

### 3.3 Direction Reversal Rules
- **[FACT]** Ghosts cannot voluntarily reverse direction; at an intersection, they can only continue straight or turn to a side, never flip 180°.
- **[FACT]** The game engine, however, forcibly reverses all ghosts' direction whenever a **mode transition** occurs between: chase→scatter, chase→frightened, scatter→chase, scatter→frightened. No forced reversal occurs when leaving frightened mode back into chase or scatter.
- **[FACT]** The reversal command is not instantaneous for every ghost simultaneously: each ghost only honors the "reverse" signal once it reaches the next tile boundary along its current path, meaning ghosts can appear to reverse at slightly different times after a mode change.
- **[DECISION]** The ghost-only no-voluntary-reversal rule does not apply to Pac-Man: opposite directional input reverses his direction immediately, including in a corridor.
- **Design implication:** mode transitions must be modeled as discrete events that set a "pending reversal" flag per ghost, consumed at the next tile-entry check, rather than mutating ghost direction immediately and globally.

### 3.4 Frightened Mode Pathing
- **[FACT]** While frightened, ghosts abandon target-based pathing and instead use a pseudo-random choice at each intersection: a PRNG-derived direction is tried first; if blocked by a wall or opposite to current travel, the next direction is tried clockwise, repeating until a valid one is found.
- **[FACT]** The PRNG is deterministically reseeded at the start of each level and whenever a life is lost, which is why frightened-ghost paths are reproducible within a session under identical conditions ("patterns").
- **[FACT]** Bit-for-bit reproduction of the original PRNG/memory-address trick is not required for a faithful "feel"; a seeded PRNG with the same reseed events (level start, life lost) is sufficient to reproduce the qualitative behavior (deterministic-but-unpredictable wandering) without needing the original's exact bit manipulation.

### 3.5 Pac-Man's Eating Pause
- **[FACT]** Eating a regular dot halts Pac-Man's movement for 1 frame (1/60 s); eating an energizer halts movement for 3 frames. This is a deliberate ~10% average speed penalty that lets a pursuing ghost gain ground, acting as a balancing mechanism against players who eat continuously in a straight line.
- **Design implication:** this is a per-frame, per-tick effect, not a "cooldown" expressed in wall-clock time — it should be modeled as frame-counted state tied to the fixed-timestep update loop (§9.2), not real-time delays.

### 3.6 Cornering
- **[FACT]** Unlike ghosts, Pac-Man can begin turning several pixels *before* reaching the center of an intersection ("pre-turn") and can also complete a turn a few pixels *after* passing the center ("post-turn"). During a pre-/post-turn, Pac-Man moves diagonally at a 45° angle until he reaches the centerline of his new direction, at which point full-speed, single-axis movement resumes.
- **[FACT]** The number of available pre-turn/post-turn pixels differs depending on the direction from which the intersection is approached (3 pixels if Pac-Man is moving right or up vs. 4 pixels if Pac-Man is moving left or down) — see `Cornering.png` and `cornering_example2.png`, which diagram the pre-turn (green), center (yellow), and post-turn (red) pixel zones for all four approach directions.
- **Design implication:** cornering is a *player-skill* mechanic that must be reproduced as a genuine input-timing allowance (early-turn tolerance window around intersections), not merely cosmetic. It directly affects competitive play and should be modeled in the player movement/collision logic, not just the animation.

### 3.7 Fruit and Fruit Counter
- **[FACT]** Bonus fruit spawns twice per round: after 70 dots eaten, and after 170 dots eaten. It appears below the ghost house, stays on screen a randomized 9–10 seconds (not a fixed value, and not reproducible via pattern-replay), and is worth a level-dependent amount (see Table in §5.3).
- **[FACT]** A "fruit/level counter" strip shows the symbols for the last six completed rounds plus the current one, along the bottom of the screen.
- **[ASSUMPTION]** Chosen HUD layout: the current level symbol occupies the far-right slot, mirrored horizontally from the leftmost reserve-life icon and at the same height. Earlier levels occupy slots every two grid columns to its left. On advancing, the new symbol enters the rightmost slot and the previous symbols shift left, dropping the oldest when more than seven would be visible. Repeated symbols remain separate icons; eating or missing a bonus fruit does not change this counter.
- **[DECISION]** The eaten-dot count that triggers the fruit includes energizers (244 collectibles total), so the first fruit appears when 174 collectibles remain and the second when 74 remain. Each trigger fires exactly once per level because the count only grows.
- **[DECISION]** The fruit is centered horizontally on the screen (between columns 13 and 14) and vertically on zero-based row 20, the corridor directly below the ghost house; this is the same horizontal position as Pac-Man's start.
- **[DECISION]** The visibility duration is drawn uniformly from 540 to 600 ticks inclusive (9–10 s at 60 Hz) from a pseudo-random generator that is independent of the frightened-mode PRNG and is not reseeded at level start or life loss, so fruit timing is not pattern-reproducible, matching the source. The generator is injectable for deterministic tests.
- **[DECISION]** Pac-Man eats the fruit when his center point lies within half a cell (8 logical pixels) of the fruit's center on both axes, i.e., when the sprites actually overlap, rather than by tile occupancy. Eating the fruit pauses Pac-Man for 1 frame, like a regular dot.
- **[DECISION]** After the fruit is eaten, its point value replaces it for 2 seconds (120 ticks), rendered in pink with the HUD font because the sprite atlas has no numeric score sprites. The popup is purely visual.
- **[DECISION]** A visible fruit or score popup remains frozen during the death collision freeze and disappears when the death animation begins; it is not rescheduled. It disappears immediately on level clear. Fruit timers advance only during active play.

---

## 4. Ghost Behavior, States, and Decision Logic

### 4.1 The Three Modes
| Mode           | Trigger                                                                        | Ghost objective                                                              | Visual cue                                   |
| -------------- | ------------------------------------------------------------------------------ | ---------------------------------------------------------------------------- | -------------------------------------------- |
| **Chase**      | Timer-driven (alternates with scatter); default/indefinite after the 4th cycle | Target tile determined by ghost-specific function (often Pac-Man-relative)   | Normal color per ghost                       |
| **Scatter**    | Timer-driven, 4 occurrences per level                                          | Head toward a fixed, unreachable "home corner" tile in dead space, then loop | Normal color per ghost                       |
| **Frightened** | Pac-Man eats an energizer                                                      | No target; pseudo-random wandering, vulnerable to being eaten                | Dark blue, then flashing white before expiry |

- **[FACT]** All three modes share identical movement/pathfinding mechanics (lookahead + target tile), except frightened mode replaces target-seeking with the PRNG-driven wandering described in §3.4.
- **[FACT]** Mode changes force a direction reversal (§3.3), which is the main *visible* signal of a mode switch to the player.

### 4.2 Scatter/Chase Timing Table
**[FACT]** (from the dossier's table, values in seconds):

| Mode cycle | Level 1    | Levels 2–4 | Levels 5+  |
| ---------- | ---------- | ---------- | ---------- |
| Scatter 1  | 7          | 7          | 5          |
| Chase 1    | 20         | 20         | 20         |
| Scatter 2  | 7          | 7          | 5          |
| Chase 2    | 20         | 20         | 20         |
| Scatter 3  | 5          | 5          | 5          |
| Chase 3    | 20         | 1033       | 1037       |
| Scatter 4  | 5          | 1/60       | 1/60       |
| Chase 4    | indefinite | indefinite | indefinite |

- **[FACT]** This scatter/chase timer is **paused** while any frightened period is active, resuming exactly where it left off once frightened mode ends.
- **[FACT]** The timer resets to the beginning of the cycle (restarting at Scatter 1) whenever a life is lost or a level is completed.
- **[FACT]** The 4th scatter period is reduced to a near-instantaneous 1/60 s on levels 2+, which in practice looks like a simple direction reversal rather than an actual scatter excursion.
- **Design implication:** this is best modeled as a per-level, ordered sequence of (mode, duration) entries driven by a single countdown timer shared by all ghosts (not per-ghost), with pause/resume semantics tied to the frightened state.

### 4.3 Target-Tile Logic — Shared Pathfinding
- **[FACT]** All ghosts share one algorithm (§3.2): on entering a tile, look ahead one tile; at the next intersection, discard the reverse direction, evaluate remaining exits by distance-to-target of the tile one step beyond, break ties Up > Left > Down > Right.
- **[FACT]** The *only* behavioral difference between the four ghosts (and between chase and scatter mode) is the computation of the **target tile** itself.

### 4.4 Per-Ghost Target Computation (Chase Mode)

**Blinky ("Shadow")** — **[FACT]**
- Target tile = Pac-Man's current tile, directly. The simplest and most aggressive scheme.
- Special "**Cruise Elroy**" behavior: on each level, once remaining dots fall at/below a level-defined threshold ("Elroy 1"), Blinky speeds up to at least Pac-Man's speed; at a second, lower threshold ("Elroy 2"), he speeds up again to faster than Pac-Man. While in Elroy state, Blinky's *scatter* target also switches to Pac-Man's tile (i.e., he keeps chasing even during scatter periods), though he still obeys forced reversals at mode changes.
- **[FACT]** Elroy status is suspended after Pac-Man loses a life, until Clyde is seen exiting the ghost house again; only then does Elroy tracking resume based on the current dot count.
- **[DECISION]** Before Clyde was implemented, Elroy remained suspended for the rest of the level after a life was lost. With Clyde present, restore the normal Clyde-exit rule instead.
- See `Blinky.png` for the direct-targeting illustration.

**Pinky ("Speedy")** — **[FACT]**
- Target tile = 4 tiles ahead of Pac-Man in his current direction of travel.
- **[FACT] Known bug (preserved by the dossier as authentic behavior):** when Pac-Man is moving **up**, the target is computed as 4 tiles up **and** 4 tiles to the left, due to an overflow bug in the original logic. This is documented, not an error to silently "fix" — see `Pinky1.png`/`Pinky2.png`.
- This scheme makes Pinky exploitable via direction changes ("head faking"): a quick reversal just before Pinky reaches an intersection can redirect him away, because his target tile moves behind him.

**Inky ("Bashful")** — **[FACT]**
- Target tile is derived from **both** Pac-Man's position/orientation and Blinky's current tile: compute an intermediate offset 2 tiles ahead of Pac-Man (in his facing direction), then draw a vector from Blinky's tile through that offset tile and double its length — the tile reached is Inky's target.
- **[FACT]** The same upward-movement overflow bug affects the intermediate offset calculation: when Pac-Man faces up, the offset is computed as 2 tiles up **and** 2 tiles left, not purely up. See `InkyA.png` / `InkyB.png`.
- Because the target depends on Blinky's position, Inky's apparent "personality" shifts: predictable/evasive when Blinky is far from Pac-Man, aggressive when Blinky is close.

**Clyde ("Pokey")** — **[FACT]**
- Compute Euclidean distance between Clyde and Pac-Man's tiles. If ≥ 8 tiles, target Pac-Man directly (same as Blinky). If < 8 tiles, target switches to Clyde's own scatter-mode corner tile instead, i.e., he "retreats" until he's far enough away, then starts chasing again — producing a looping approach/retreat pattern near Pac-Man. See `ClydeJ.png`.
- **[AMBIGUITY]** The dossier does not state whether the 8-tile threshold check happens continuously every tile-entry or only at intersections. **[ASSUMPTION]** Evaluate it on every tile-entry (same granularity as the "look ahead" step in §3.2), since Clyde's target can conceptually flip at any point along a straight corridor, not only at junctions.

### 4.5 Fixed Target Tiles (Scatter Mode & "Go Home" State)
- **[FACT]** Each ghost has one fixed scatter target in dead space outside the maze proper, located near its respective home corner — unreachable by design, which is why ghosts circle their corner rather than parking on a single tile. See `Scatter.png`.
- **[FACT]** A separate fixed target (just above/left of the ghost-house door) is used when a ghost has been eaten and its "eyes" are returning home to be revived — this is effectively a 4th pseudo-mode ("eaten"/"go home"), using the same pathfinding mechanism with yet another fixed target.
- **Design implication:** model ghost mode as **4 states**, not 3: Chase, Scatter, Frightened, and Eaten (returning to pen). "Eaten" is implied by the dossier's description of eyes needing a target to navigate home — it must be a distinct state in the state machine even if the dossier groups it loosely under "frightened aftermath."

### 4.6 Red Zones and Tunnel Zones (Topology Constraints)
- **[FACT]** Four "red zones" (upward-tunnel entrances) exist where ghosts are forbidden from turning upward during chase/scatter — only Pac-Man can use these as shortcuts; ghosts can still enter them from the opposite end, however. In frightened mode, this restriction is lifted and ghosts may turn upward there. See `exploit.png`.
- **[FACT]** "Pink zones" are the two tunnel-interior stretches where any ghost's speed is cut nearly in half; Pac-Man is unaffected by this slow-down.
- **Design implication:** these are **tile-tagged behavioral rules**, not geometry exceptions — the maze data model should support per-tile flags (e.g., `NoGhostUpturn`, `GhostSlowZone`) evaluated by the ghost movement/decision system, independent of mode.

---

## 5. Levels, Difficulty Progression, Scoring, and Lives

### 5.1 Speed Progression (Pac-Man & Ghosts)

| Level | Pac-Man Speed | Ghost Speed (normal) |
| ----- | ------------- | -------------------- |
| 1     | 80%           | 75%                  |
| 2–4   | 90%           | 85%                  |
| 5–20  | 100%          | 95%                  |
| 21+   | 90%           | 95%                  |

- **[FACT]** "100% speed" is quantified in the dossier as 75.75757625 pixels/second — a concrete reference value worth preserving for faithful timing even if the clone's resolution/scale differs (scale all speeds relative to tile size, not absolute pixels, if the grid size changes).
- **[DECISION]** Based on the project owner's arcade reference, returning eyes use 200% of that base speed at every level, including in the side tunnels: approximately 151.5 original pixels/second, or 18.94 tiles/second. They retain the fixed house target and shared navigation routine.
- **[FACT]** Ghosts in the tunnel move at a further-reduced "tunnel speed" (e.g., 40–50% depending on level — see Appendix table).
- **[FACT]** From level 21 onward, ghosts move *faster* than Pac-Man even at normal (non-Elroy) speed — a notable difficulty inflection point.

### 5.2 Frightened Mode Duration & Flashes
- **[FACT]** Frightened duration and flash count shrink across levels (down to as little as 1 second with 3 flashes on later levels), and **from level 19 onward ghosts no longer turn blue / cannot be eaten at all** — they still reverse direction on the energizer trigger, but there is no vulnerability window. Appendix Table A.1 in the dossier gives exact per-level seconds and flash counts.
- **[AMBIGUITY]** The dossier states ghosts stop turning blue "by level 19" in the main text, but Table A.1 shows levels 17, 19, 20, and 21+ with "–" (no fright data) while level 18 still lists 1 second / 3 flashes. **[ASSUMPTION]** Treat the per-level table as authoritative over the prose summary: fright mode is effectively disabled (no color change, no vulnerability, ghosts simply reverse) specifically on levels 17, 19, 20, and 21+, and still active (if very short) on level 18. Implement per-level fright duration as data-driven from the table, not from the rounded prose statement.

### 5.3 Bonus Fruit Table (per level)

| Level | Symbol     | Points |
| ----- | ---------- | ------ |
| 1     | Cherries   | 100    |
| 2     | Strawberry | 300    |
| 3–4   | Peach      | 500    |
| 5–6   | Apple      | 700    |
| 7–8   | Grapes     | 1000   |
| 9–10  | Galaxian   | 2000   |
| 11–12 | Bell       | 3000   |
| 13+   | Key        | 5000   |

**[FACT]**, directly from Appendix Table A.1.

### 5.4 Elroy Thresholds (Blinky Speed-Up)
- **[FACT]** Each level defines two dot-count thresholds ("Elroy 1" and "Elroy 2" dots-remaining) at which Blinky accelerates, growing larger (i.e., triggering earlier, with more dots left) as levels progress (e.g., level 1: 20/10 dots remaining; level 19+: 120/60). Full per-level values are in Appendix Table A.1 of the source dossier and should be carried into the clone's level-configuration data verbatim.

### 5.6 Lives
- **[FACT]** Pac-Man start a game with 3 lives, when lives end the game is over.
- **[FACT]** Losing a life resets: Pac-Man and ghosts to start positions, the scatter/chase timer to its beginning, ghosts' per-ghost dot counters to disabled (replaced temporarily by the global dot counter, §6.2), and the frightened PRNG seed.
- **[DECISION]** A match starts with three lives and awards no extra lives for reaching a score threshold. The source dossier does not specify a threshold; this project keeps the current behavior.

---

## 6. Ghost House Logic (Spawn/Release Rules)

This is one of the most detail-dense subsystems in the dossier and deserves explicit, careful treatment since subtle bugs here are easy to introduce.

### 6.1 Starting Positions
- **[FACT]** Blinky starts just outside, above the ghost house door. Inky, Pinky, Clyde start inside (left, middle, right respectively). See `StartPositions.png`.
- **[FACT]** Once a ghost leaves the house, it cannot re-enter except as "eyes" after being eaten by Pac-Man, in which case it travels to the fixed "return home" target tile (§4.5), re-enters, and is revived (then exits again, resetting to normal mode).
- **[DECISION]** Inky and Clyde revive in their original left and right house positions, respectively, after returning as eyes; the existing personal/global release rules still determine when they exit again.
- **[FACT]** Because Blinky starts outside, he is functionally never subject to the house-release logic below except after being eaten.
- **[FACT]** Ghosts waiting in the house bounce vertically; the source dossier specifically describes Clyde bouncing until he moves toward the exit.
- **[DECISION]** Use a fixed 40% of base speed for this bounce at every level and in every mode, as chosen by the project owner. The implementation uses center positions from rows 17 to 18 in the 16-pixel grid and begins exiting immediately from the current height. Both the speed and bounds are approximations, not verified pixel-exact arcade behavior.

### 6.2 Release Conditions (Pinky, Inky, Clyde)
Two independent, concurrently-active controls govern when a ghost leaves the house:

**A. Per-ghost dot counters (level-start condition)**
- **[FACT]** Each of the 3 ghosts has a personal counter and a level-specific dot limit. Only one ghost's counter is "active" at a time, in preference order Pinky > Inky > Clyde; as soon as the active ghost reaches/exceeds its limit, it exits and the next most-preferred waiting ghost's counter activates.
- **[FACT]** Limits by level: Pinky's limit is always 0 (leaves immediately). Level 1: Inky = 30, Clyde = 60. Level 2: Inky = 0, Clyde = 50. Level 3+: all three = 0 (all leave immediately at level start).
- **[DECISION]** Preserve the arcade release rules after Inky or Clyde revives, rather than always sending them out immediately. In particular, if Inky was forced out on level 1 before reaching 30 dots, his retained personal count resumes when he returns; the anti-stall timer can still release him.

**B. Global dot counter (post-death condition)**
- **[FACT]** After a life is lost, per-ghost counters are disabled (not reset) and replaced by a single global counter, reset to 0, counting dots eaten since the respawn. Pinky releases at count = 7, Inky at count = 17. At count = 32, the global counter resets and deactivates only if Clyde is still waiting in the house; if Clyde has already started leaving (e.g., forced out by the timer, below), the counter remains active and keeps counting indefinitely, with no further release checkpoints defined.
- **[DECISION]** At 32 with Clyde waiting, re-enable the retained personal counters and apply their normal limits and priority; Clyde does not automatically leave unless his personal limit permits it. This follows the source dossier's description and resolves the earlier project-dossier wording that said Clyde always leaves at 32.
- **[DECISION]** Before Clyde was implemented, the global counter remained active after a life was lost rather than simulating his release at 32. With Clyde present, use the presence rule above.

**C. "No-dot-eaten" timer (anti-stalling safeguard)**
- **[FACT]** A continuously running timer, reset to 0 every time a dot is eaten, forces the most-preferred waiting ghost out immediately if it ever reaches its limit (4 seconds initially, 3 seconds from level 5 onward) — this prevents a player from indefinitely trapping ghosts by refusing to eat dots.

- **[FACT] Known exploit chain** (explicitly documented in the dossier, not a design error): deliberately starving the no-dot timer to force Clyde out *before* the global counter reaches 32 permanently disables the global counter's ability to release ghosts via counting, leaving only the anti-stall timer as a release mechanism — allowing a skilled player to keep Pinky/Inky/Clyde "stuck" in the house for an extended period after they've been eaten, as long as dots are eaten frequently enough to keep resetting the stall timer.
- **[AMBIGUITY]** The dossier presents this exploit as a *known, authentic quirk* of the original game rather than specifying whether a clone should preserve it. **[ASSUMPTION]** Preserve it by implementing the rules exactly as specified (not by special-casing "block the exploit"), since the brief's goal is behavioral fidelity to the original; the exploit is an emergent property of correctly implementing the stated rules, not a separate feature to build.

### 6.3 Exit Direction (Left vs. Right)
- **[FACT]** A ghost leaving the house normally turns left on exit, but if one or more mode changes (chase↔scatter) occurred while it was still inside, it turns right instead.
- **Design implication:** each ghost needs a per-stay "mode-change-occurred-while-inside" flag, checked once at the moment of exiting and then cleared.

---

## 7. Maze Topology, Tiles, and Collision Edge Cases

### 7.1 Tile Grid Model
- **[FACT]** 8×8-pixel tiles; 28×36 grid for the canonical 224×288 screen. See `Tiles.png` for the legal-space/dead-space layout, and `TileGame.png` for the abstracted tile/direction view that the AI logic actually operates on.
- **[ASSUMPTION]** A clone may use a different pixel-per-tile scale (e.g., for higher-resolution displays), but the **tile-based reasoning model itself (8-unit grid cells, actor position = center-point tile)** must be preserved structurally, since all ghost AI, dot placement, and distance calculations are tile-relative in the source material.
- **[DECISION]** Tunnel traversal is visually continuous and symmetric for Pac-Man and every ghost. An actor may extend beyond the exiting screen edge while crossing; render complementary portions of the same sprite at both edges until its center wraps to the opposite side. Do not teleport the full sprite across the screen, overlap the two portions, or leave a visual gap. Preserve the ghost-only tunnel speed penalty.

### 7.2 Actor-Tile Association
- **[FACT]** An actor's tile is determined purely by its center point, updated continuously as pixel position crosses tile boundaries — not snapped or rounded at intervals. See `TileMove2.png` for a frame-by-frame illustration of this update as a ghost crosses a tile boundary.

### 7.3 The "Pass-Through" Collision Edge Case
- **[FACT]** Because collision is checked as "same tile," there's a known, rare edge case where Pac-Man and a ghost moving toward each other swap tiles within the same single simulation frame (each one's center crosses into the other's tile simultaneously), resulting in no frame where they share a tile — hence no detected collision, even though they visually overlapped. See `PassthruBug.png` for the 5-frame sequence.
- **[ASSUMPTION]** This is explicitly framed by the dossier as an authentic emergent property of discrete, tile-based, per-frame collision checking — **not a bug to patch**. The correct implementation approach is to use the exact same simple, discrete "same tile this frame" collision rule rather than continuous/swept collision detection; doing so will reproduce this behavior naturally rather than requiring it to be special-cased.

### 7.4 Intersections and the "Four Exits" Model
- **[FACT]** Every tile conceptually has 4 potential exits (up/down/left/right); walls prune the unavailable ones. Ghost decision-making (§3.2) always starts from this 4-exit model and discards invalid directions (walls, reversal) before comparing remaining candidates by distance. See `IntersectALL.png` for the two-stage illustration (test-tile collection, then distance comparison) and `TieBreakALL.png` for the tie-break example.

---

## 8. Validation Checklist Against Original Behavior

Before considering the clone "faithful," verify each of the following observable behaviors, grouped by subsystem. These are the testable, player-visible consequences of the rules above and are the most efficient way to catch subtle logic drift during development.

Checked items are implemented and covered by the headless simulation tests in `game/` (see the implementation reference for the test command); intentional audio deviations are documented in §1.10.

**Movement & Collision**
- [x] Pac-Man can begin a turn before reaching an intersection's center (cornering) and ghosts cannot.
- [x] Pac-Man and ghosts cross both tunnel edges continuously in either direction, showing complementary sprite portions at opposite edges without a jump, overlap, or gap; ghosts retain their tunnel speed penalty.
- [x] A rare, reproducible "pass-through" is observable under the exact tile-swap conditions, and only there.
- [x] Ghosts moving through the tunnel visibly slow down; Pac-Man does not.
- [x] Ghosts cannot turn upward in the four designated no-upturn tiles (except while frightened).

**Ghost AI**
- [x] Blinky always beelines for Pac-Man's current tile.
- [x] Pinky targets 4 tiles ahead of Pac-Man, with the "up = up+left" bug reproduced, not corrected.
- [x] Inky's target visibly depends on Blinky's position (verify by moving Blinky far away vs. near Pac-Man and observing Inky's path change), including the equivalent "up" bug in his offset step.
- [x] Clyde switches between direct pursuit and corner-retreat exactly at the 8-tile distance boundary, producing a visible looping pattern near that boundary.
- [x] Tie-break order (Up > Left > Down > Right) is observable when two paths are equidistant.
- [x] Ghosts never reverse direction except at verified mode-change moments.

**Modes & Timing**
- [x] Scatter/chase cycle lengths match the level-based table (§4.2), including the near-instant 4th scatter period on level 2+.
- [x] Scatter/chase timer visibly pauses during frightened periods and resumes correctly afterward.
- [x] Frightened duration and flash count shrink per the per-level table, and fright mode visibly disappears (no blue, no vulnerability) on levels 17/19/20/21+ while level 18 retains a brief fright window (per the table-vs-prose resolution in §5.2).
- [x] Cruise Elroy triggers at the correct dot-remaining thresholds, speeds Blinky appropriately, and is suspended after a death until Clyde is seen leaving the house again.

**Level Transitions**
- [x] Clearing the final collectible starts a 120-tick (2-second) maze flash that alternates blue and white four times; Pac-Man stops in place and all ghosts disappear immediately.
- [x] The next level shows a centered yellow `READY!` for approximately 2 seconds, with Pac-Man and all ghosts visible, stationary, and at their starting positions. The same intro appears at match start and after a life is lost.
- [x] The start jingle plays during a match's opening intro; controls remain locked and actors remain still until both `READY!` has disappeared and the jingle has ended. Later intros are silent and shorter.

**Ghost House**
- [x] At level start, Pinky/Inky/Clyde leave in the documented order and timing for levels 1, 2, and 3+ respectively.
- [x] After a death, the global counter governs Pinky's and Inky's releases at 7/17; at 32 the Clyde-presence condition and restoration of personal counters behave as documented.
- [x] The anti-stall timer (4s/3s) forces a release if dots go uneaten too long.
- [x] Ghosts exit left normally, right if a mode change happened while they waited inside.

**Scoring & Progression**
- [x] Ghost-eating streak scores 200/400/800/1600, resetting with each new energizer.
- [x] Bonus fruit appears at 70 and 170 dots eaten, lasts a randomized 9–10 seconds, and matches the level's symbol/value table.
- [x] Fruit/level history strip shows the last 6 rounds + current.
- [x] Speed percentages for Pac-Man and ghosts match the per-level bands (80/90/100/90%).

---

## 9. Architectural Guidance for Go + Ebiten

This section stays at the level of **concepts and responsibilities**, deliberately avoiding code, package layouts, or function signatures, per the brief's constraints.

### 9.1 Separation of Concerns
Three largely independent concerns emerge directly from the dossier's own structure, and should map to analogous separations in the implementation:

1. **Simulation / rules layer** — tile grid, actor tile-position tracking, mode state machines, timers, scoring, dot counters, ghost target-tile computation. This layer should be *renderer-agnostic*: it reasons purely in tiles, frames, and game state, with no dependency on Ebiten's drawing primitives. This mirrors the dossier's own framing of the AI as operating on an abstract tile model (`TileGame.png`) regardless of pixel-level visuals.
2. **Presentation / rendering layer** — sprites, animations, screen-space positioning (derived from the simulation's sub-tile pixel position), the fruit/level counter strip, score/lives HUD, maze tile artwork. This layer reads simulation state but never mutates it.
3. **Input / scene layer** — menu, "ready" countdown, gameplay scene, death animation, game-over scene, pause. Think of this as a small scene/state stack sitting above the simulation and renderer, each scene owning which simulation/renderer calls are active and which input is accepted.

### 9.2 Game Loop and Timestep
- **[FACT]-grounded design driver:** virtually every rule in the dossier is expressed in discrete frames (1/60 s) — the eating pause (1 or 3 frames), the one-frame reversal response latency, the 1/60 s fourth scatter period, PRNG reseeding events, etc. This strongly implies a **fixed-timestep simulation update**, decoupled from Ebiten's rendering callback, so that gameplay logic always advances in consistent, countable "ticks" rather than being tied to variable frame rendering time.
- Ebiten's standard `Update`/`Draw` separation maps naturally onto this: `Update` should drive the simulation at a fixed logical rate (ideally matching the original's 60 Hz tick granularity so that frame-counted rules like the eating pause translate directly), while `Draw` is purely a read-only projection of current simulation state into pixels, including any pixel-level cornering interpolation.
- **Design implication:** avoid delta-time-scaled movement for core logic (dot-eating pause, mode timers, PRNG reseed points); keep these as integer tick counts. Only purely cosmetic interpolation (e.g., smoothing animation) should use continuous/delta time.

### 9.3 State Machines as the Core Abstraction
Two layered state machines recur throughout the dossier and should be first-class concepts in the architecture, not incidental flags:

- **Global game-flow state machine:** attract/menu → ready countdown → playing → (level complete? 2-second level-clear flash → next-level READY/jingle → playing : death sequence → life remaining? ready countdown : game over) → high-score/attract. Level-clear and new-level intro timing/visibility are specified in §1.6.
- **Per-ghost mode state machine:** Chase ⇄ Scatter (timer-driven, see §4.2) with Frightened and Eaten as interrupting states that can be entered from either and that resume the interrupted state's timer afterward (§4.2, §4.5). Each ghost instance needs its own mode state, but the **chase/scatter timer itself is shared/global**, not per-ghost — this distinction matters architecturally: don't model it as four independent timers that could drift out of sync.

### 9.4 Target-Tile Strategy as a Pluggable Responsibility
Because the dossier establishes that **all four ghosts share one pathfinding routine** and differ only in **how they compute a target tile** (§4.3–4.4), the architecture should reflect this directly: a single, shared "navigate toward target tile" responsibility (lookahead, exit evaluation, tie-break) parameterized by a per-ghost, per-mode "what is my target tile right now" responsibility. This keeps the four ghosts' unique personalities isolated to one well-defined extension point rather than duplicated across four separate movement implementations — directly reflecting how the source material itself explains the AI design.

### 9.5 Data-Driven Level Configuration
Nearly every numeric rule in the dossier (speeds, fright duration/flashes, Elroy thresholds, fruit symbol/value, dot-release limits) is **indexed by level number**, with values stabilizing after a certain level (21+ in most tables). This should be modeled as a lookup table/data structure keyed by level, not as a tree of conditional branches in code — matching the dossier's own presentation as Appendix tables. Resolve ambiguities such as §5.2 (table vs. prose for when fright mode disables) by treating the table as the source of truth and deriving behavior from it directly, so that updating one row of configuration data is enough to adjust a level's difficulty.

### 9.6 Maze/Tile Data Model Responsibilities
The maze itself should own:
- Legal/dead space per tile (navigability).
- Special per-tile flags: no-ghost-upturn ("red zones"), ghost-slow zone ("pink/tunnel zones"), tunnel wrap pairs, ghost-house door, ghost-house interior, dot/energizer/fruit spawn points, and the four fixed scatter corner targets (which live in dead space, outside normal navigable tiles, by design — §4.5).
- This mirrors the dossier's treatment of the maze as a single static grid annotated with behavioral meta-information, rather than purely visual geometry.

### 9.7 Determinism and Testability
Because so much of the ghost behavior described in the dossier is precisely specified (exact distances, exact tie-break order, exact timer values, exact dot-counter thresholds), the simulation layer should be designed for **deterministic, headless execution** — i.e., it must be possible to run many simulated ticks without the Ebiten renderer attached, feeding scripted input, in order to verify the checklist in §8 automatically (e.g., asserting that Clyde's target switches exactly at the 8-tile boundary, or that the global dot counter releases Pinky exactly at 7). This argues strongly for keeping the simulation layer free of any direct Ebiten calls, consistent with §9.1.

---

## 10. Summary of Open Items Requiring a Decision

| #   | Topic                                | Dossier status                                                           | Recommended default                                                                                                                                   |
| --- | ------------------------------------ | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Exact maze layout                    | Only aggregate stats given (244 dots, 4 energizers, tunnel, 4 red zones) | Design a custom maze respecting the same topology constraints and dot totals (§1.2)                                                                   |
| 2   | Starting life count                  | Not stated in this excerpt                                               | 3 lives, configurable (§5.6)                                                                                                                          |
| 3   | Fright mode cutoff (levels 17–21)    | Table and prose slightly disagree                                        | Follow Appendix Table A.1 per level, not the rounded prose statement (§5.2)                                                                           |
| 4   | Hard-mode jumper levels              | Documented as a hardware curiosity, not a core rule                      | Treat as optional/stretch feature, data-driven (§5.5)                                                                                                 |
| 5   | Original PRNG bit-level reproduction | Described conceptually, not fully specified                              | Use a seeded PRNG with equivalent reseed events; exact bit-level parity not required unless frame-perfect "pattern" replay is an explicit goal (§3.4) |
| 6   | Clyde's 8-tile check granularity     | Not stated whether continuous or per-intersection                        | Evaluate every tile-entry, for responsiveness consistent with the rest of the lookahead model (§4.4)                                                  |
| 7   | Attract/title screen                 | Named in the game-flow outline (§9.3) but not specified                  | Decided (§1.9): attract presentation implemented without demo play; `Enter` replaces the coin insert and starts a single-player match; `GAME OVER` returns to the attract screen after five seconds |

---

### Image Reference Index
The following source images were consulted to resolve or illustrate specific mechanics discussed above; they are referenced by filename for traceability back to the dossier's visual material rather than embedded, since this document is a textual specification:

- `lvl1.png` — overall maze layout reference (§1.2).
- `Tiles.png`, `TileGame.png`, `TileMove2.png` — tile grid model and actor-tile association (§7.1–7.2).
- `Cornering.png`, `cornering_example2.png` — pre-turn/post-turn pixel zones (§3.6).
- `StartPositions.png` — ghost house starting layout (§6.1).
- `exploit.png` — red zone / tunnel zone map (§4.6).
- `IntersectALL.png`, `TieBreakALL.png` — intersection evaluation and tie-break example (§3.2, §7.4).
- `Scatter.png` — fixed scatter-mode target tiles per ghost (§4.5).
- `Blinky.png` — direct-targeting illustration (§4.4).
- `Pinky1.png`, `Pinky2.png` — offset targeting and the up-direction bug, plus head-fake exploit (§4.4).
- `InkyA.png`, `InkyB.png` — vector-based targeting and its own up-direction bug (§4.4).
- `ClydeJ.png` — distance-threshold targeting loop (§4.4).
- `PassthruBug.png` — frame-by-frame pass-through collision edge case (§7.3).
