# Pac-Man Clone — Go + Ebiten Implementation Proposal

## 1. Purpose and Scope

This document converts the functional design dossier into a lower-level, implementation-oriented proposal for a deterministic Pac-Man-style game built in **Go** with **Ebiten**.

The proposal prioritizes behavioral fidelity for:

- 60 Hz, frame-counted simulation.
- Tile-driven ghost navigation and collision rules.
- Pixel/sub-tile movement and Pac-Man cornering.
- Shared chase/scatter timing with per-ghost mode state.
- Data-driven level progression, maze metadata, speeds, and release rules.
- Renderer-independent simulation suitable for automated tests.

It intentionally does **not** provide production code. Pseudocode appears only for rules where sequencing is critical.

---

## 2. Design Goals

| Goal | Proposed approach |
|---|---|
| Faithful mechanics | Fixed 60 Hz simulation; integer tick counters; no delta-time game rules. |
| Deterministic behavior | Explicit seedable random generator; deterministic update ordering; immutable configuration data. |
| Testability | `sim` package has no Ebiten dependency and can run headlessly. |
| Extensibility | Maze, level configuration, sprite definitions, and scene content remain data-driven. |
| Maintainability | Separate simulation, application orchestration, input, rendering, audio, and asset loading. |

### Non-goals for the first playable milestone

- Exact arcade ROM emulation.
- Networked multiplayer.
- Runtime maze editor.
- Pixel-perfect reproduction of the original sound hardware.
- General-purpose ECS framework; the game has a small, stable set of entities and benefits from explicit domain types.

---

## 3. Architectural Overview

The recommended architecture is a **layered, explicit domain model**, with a unidirectional update flow.

```text
Ebiten Game Loop
      |
      v
Application / Scene Coordinator
      |
      +--> Input Adapter ----> InputFrame
      |
      +--> Simulation Engine -> GameState
      |                             |
      |                             +--> Events
      |
      +--> Renderer <--------------+ 
      |
      +--> Audio Controller <------ Events
```

### Core rule

The simulation must not import Ebiten. Ebiten is limited to the application, renderer, input, audio, and asset packages.

This ensures that tests can create a game state, submit input frames, advance a chosen number of ticks, and assert the resulting state without opening a window or loading graphic assets.

---

## 4. Proposed Project Layout

```text
pacman-clone/
├── cmd/
│   └── pacman/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── game.go
│   │   ├── scene_manager.go
│   │   └── config.go
│   ├── sim/
│   │   ├── engine.go
│   │   ├── game_state.go
│   │   ├── phase.go
│   │   ├── events.go
│   │   ├── player.go
│   │   ├── ghost.go
│   │   ├── ghost_ai.go
│   │   ├── ghost_house.go
│   │   ├── mode_scheduler.go
│   │   ├── movement.go
│   │   ├── collision.go
│   │   ├── collectibles.go
│   │   ├── fruit.go
│   │   ├── scoring.go
│   │   ├── level_rules.go
│   │   ├── random.go
│   │   └── maze/
│   │       ├── maze.go
│   │       ├── tile.go
│   │       ├── topology.go
│   │       └── canonical_maze.go
│   ├── input/
│   │   ├── adapter.go
│   │   └── bindings.go
│   ├── render/
│   │   ├── renderer.go
│   │   ├── world_renderer.go
│   │   ├── hud_renderer.go
│   │   ├── sprite_renderer.go
│   │   ├── animation.go
│   │   └── palette.go
│   ├── audio/
│   │   ├── controller.go
│   │   └── sound_bank.go
│   ├── assets/
│   │   ├── loader.go
│   │   ├── sprites.go
│   │   └── fonts.go
│   └── testkit/
│       ├── scenario.go
│       ├── scripted_input.go
│       └── assertions.go
├── assets/
│   ├── images/
│   ├── sounds/
│   ├── fonts/
│   ├── maze/
│   │   └── canonical_maze.json
│   └── config/
│       └── levels.json
├── tests/
│   ├── sim/
│   └── integration/
├── go.mod
└── README.md
```

### Package responsibilities

| Package | Responsibility | Must not do |
|---|---|---|
| `sim` | All gameplay rules and deterministic state transitions. | Access keyboard, draw sprites, play sound. |
| `sim/maze` | Topology, tile tags, portals, spawn locations, collectible placement. | Know UI or animation details. |
| `app` | Ebiten lifecycle, scene sequencing, dependency wiring. | Reimplement gameplay rules. |
| `input` | Convert device state into game-neutral commands. | Modify actors directly. |
| `render` | Convert read-only simulation state into visuals. | Mutate simulation state. |
| `audio` | Translate simulation events into sound playback. | Decide game rules. |
| `assets` | Load and validate runtime resources. | Contain gameplay decisions. |
| `testkit` | Build repeatable headless scenarios. | Be required by production runtime. |

---

## 5. Main Domain Types

Go has no classes; the design should use **structs, typed enums, interfaces only at stable extension points, and package-level constructors**. Avoid excessive interfaces for simple data holders.

### 5.1 Primitive value types

```text
TilePos        Grid position: row, column.
PixelPos       Fixed-point world position.
Direction      None, Up, Left, Down, Right.
TileFlags      Bit flags describing tile behavior.
Tick           Integer simulation tick.
LevelNumber    Positive integer level index.
ActorID        Player or one of the four ghosts.
```

### 5.2 Why fixed-point positions are recommended

Store moving positions in a fixed-point unit rather than `float64`.

Recommended convention:

- `1 tile = 8 logical pixels`.
- `1 logical pixel = 256 sub-units`.
- Position is stored as signed integer sub-units.
- Movement speed is stored as sub-units per tick.

Benefits:

- Exact, repeatable tile-boundary checks.
- No floating-point drift after long sessions.
- Precise representation of fractional per-tick speeds.
- Reproducible collision and turn timing across platforms.

A rendering scale can enlarge logical pixels for modern displays without changing simulation coordinates.

### 5.3 Actor model

```text
ActorMotion
  Position              PixelPos
  Direction             Direction
  NextDirection         Direction
  PendingReverse        bool
  CurrentTile           TilePos
  PreviousTile          TilePos
  Speed                 FixedSpeed

PlayerState
  Motion                ActorMotion
  RequestedDirection    Direction
  EatingPauseTicks      int
  LivesRemaining        int
  Score                 int
  AnimationState        PlayerAnimationState

GhostState
  Identity              Blinky | Pinky | Inky | Clyde
  Motion                ActorMotion
  Mode                  GhostMode
  BaseMode              Chase | Scatter
  HouseState            HouseState
  FrightenedTicks       int
  FlashTicks            int
  WasModeChangedInside  bool
  ElroyStage            None | Stage1 | Stage2
  TargetTile            TilePos
```

`NextDirection` is deliberately distinct from `Direction`: ghost decisions are made ahead of the actual intersection and are committed later when the actor reaches the corresponding decision point.

---

## 6. Core State Model

### 6.1 Global simulation state

```text
GameState
  Phase                 GamePhase
  Tick                  Tick
  Level                 LevelState
  Maze                  MazeState
  Player                PlayerState
  Ghosts                [4]GhostState
  ModeScheduler         ModeSchedulerState
  HouseController       GhostHouseState
  Fruit                 FruitState
  Score                 ScoreState
  Random                RandomState
  PendingEvents         []GameEvent
```

### 6.2 Game phases

```text
Boot
Attract
Ready
Playing
PlayerDying
LifeReset
LevelComplete
GameOver
Paused
```

`GamePhase` controls which systems are active. For example:

| Phase | Movement | Collectibles | Collision | Mode timer | Input |
|---|---:|---:|---:|---:|---:|
| Ready | No | No | No | No | Direction may be buffered |
| Playing | Yes | Yes | Yes | Yes, except frightened pause | Yes |
| PlayerDying | No | No | No | No | No |
| LifeReset | No | No | No | No | Direction may be cleared |
| LevelComplete | No | No | No | No | No |

### 6.3 Ghost modes and house states

Use two separate state dimensions rather than one oversized enum.

```text
GhostMode
  Chase
  Scatter
  Frightened
  Eaten

HouseState
  Outside
  WaitingInHouse
  LeavingHouse
  ReturningHome
  Reviving
```

This prevents invalid combinations from being hidden inside ambiguous flags. `Eaten` describes behavior and rendering; `ReturningHome` describes the ghost’s physical relationship with the ghost house.

---

## 7. Maze Data Model

### 7.1 Tile representation

The maze should be authored as structured data, loaded once, validated at startup, and converted into an efficient in-memory representation.

```text
MazeDefinition
  Width, Height
  Tiles                 []TileDefinition
  Portals               []PortalPair
  PlayerSpawn           SpawnPoint
  GhostSpawns           map[GhostIdentity]SpawnPoint
  HouseDoor             TilePos
  HouseExitPath         []TilePos
  ScatterTargets        map[GhostIdentity]TilePos
  FruitSpawn            TilePos

TileDefinition
  WalkableForPlayer     bool
  WalkableForGhost      bool
  Flags                 TileFlags
  Collectible           None | Dot | Energizer
```

### 7.2 Tile flags

Use a bitset for orthogonal attributes.

```text
TileFlagGhostNoUpTurn
TileFlagGhostSlowZone
TileFlagTunnel
TileFlagGhostHouseInterior
TileFlagGhostDoor
TileFlagPlayerCorneringZone
TileFlagDecisionPoint
TileFlagFruitSpawn
```

The `DecisionPoint` flag can be precomputed during maze loading. It should not be manually authored unless a special rule requires it. A tile is normally a decision point when it has more than two valid non-reverse exits for the relevant actor class.

### 7.3 Topology API

The maze package should expose small, rule-oriented queries:

```text
CanEnter(actorKind, fromTile, direction, ghostMode) bool
NextTile(tile, direction) (TilePos, wrapped bool)
AvailableDirections(actorKind, tile, ghostMode) []Direction
HasFlag(tile, flag) bool
CollectibleAt(tile) CollectibleType
RemoveCollectible(tile)
```

The implementation must centralize portal wrapping. Movement and AI should never implement left/right tunnel logic themselves.

---

## 8. Simulation Tick Pipeline

A stable update order is essential because the specification depends on tile entry, pending reversals, collection pauses, and same-tile collision checks.

### Recommended order for a `Playing` tick

1. Capture the input command for this tick.
2. Update player direction request.
3. Update global timers that are allowed to run.
4. Resolve player movement, including eating pauses and cornering.
5. Resolve ghost movement in a stable identity order.
6. Update tile occupancy for every actor.
7. Process collectible consumption by the player.
8. Apply effects caused by collectible consumption.
9. Evaluate fruit spawn/expiry and fruit collection.
10. Resolve same-tile collisions.
11. Update ghost-house release counters and release decisions.
12. Evaluate level-completion and phase-transition conditions.
13. Emit presentation/audio events accumulated during the tick.

### Important sequencing rule

Collision must use **current tile occupancy after movement**, not continuous swept-path collision. This deliberately preserves the documented pass-through behavior when two actors exchange adjacent tiles in the same tick.

### Critical pseudocode: playing tick

```text
function AdvancePlayingTick(state, input):
    state.tick += 1
    ApplyPlayerInput(state.player, input)

    AdvanceModeSchedulerIfNotFrightened(state)
    AdvanceNoDotTimer(state.houseController)

    MovePlayer(state)
    for ghost in ghosts ordered Blinky, Pinky, Inky, Clyde:
        MoveGhost(state, ghost)

    RefreshActorTiles(state)

    collectible = ConsumeCollectibleUnderPlayer(state)
    if collectible exists:
        ApplyCollectibleEffects(state, collectible)

    UpdateFruit(state)
    ConsumeFruitIfPresent(state)

    ResolveTileCollisions(state)
    UpdateGhostHouse(state)

    if AllMazeCollectiblesConsumed(state):
        BeginLevelComplete(state)

    return state
```

---

## 9. Movement and Turning

### 9.1 Player movement

Player movement requires:

- Requested direction buffering.
- Normal centerline movement.
- Pre-turn and post-turn cornering windows.
- One-tick pause after a dot; three ticks after an energizer.
- Tunnel wrap without speed reduction.

A dedicated `PlayerMovementSystem` should own this behavior. Do not try to share it entirely with ghost movement; the entities intentionally obey different turning rules.

### 9.2 Ghost movement

Ghost movement has these distinct rules:

- Cannot voluntarily reverse.
- Applies a pending forced reversal at the next tile boundary.
- Chooses direction one tile ahead of an upcoming decision point.
- Uses normal, tunnel, frightened, eaten, and Elroy speed profiles.
- Cannot turn upward in designated tiles unless frightened.

### 9.3 Decision scheduling

To capture look-ahead behavior, maintain an explicit decision state:

```text
GhostNavigationState
  CommittedDirection     Direction
  PlannedDirection       Direction
  PlanAppliesAtTile      TilePos
```

When a ghost enters a tile, determine whether the next tile in its current direction is a decision point. If it is, compute `PlannedDirection`; when the ghost reaches that next tile, commit it.

This avoids a common implementation defect: selecting a turn only at the center of an intersection, which makes ghosts react late and changes their route behavior.

---

## 10. Ghost AI Design

### 10.1 Shared path selection service

Implement a single service responsible for non-frightened direction selection:

```text
GhostPathPlanner
  ChooseDirection(ghost, targetTile, maze) Direction
```

Its responsibilities are fixed:

1. Determine candidate exits at the planned intersection.
2. Exclude walls and the reverse direction.
3. Enforce no-up-turn restrictions for non-frightened ghosts.
4. Compute the test tile one step beyond each candidate exit.
5. Compare squared Euclidean distance to target.
6. Apply the fixed tie order: Up, Left, Down, Right.

Use **squared distances**. The ordering is identical to Euclidean distance while avoiding square roots.

### 10.2 Critical pseudocode: target-driven choice

```text
function ChooseDirection(ghost, decisionTile, target, maze):
    candidates = []

    for direction in [Up, Left, Down, Right]:
        if direction is Opposite(ghost.motion.direction):
            continue
        if not maze.CanEnter(Ghost, decisionTile, direction, ghost.mode):
            continue

        testTile = maze.NextTile(decisionTile, direction)
        distance2 = SquaredDistance(testTile, target)
        candidates.append(direction, distance2)

    return candidate with smallest distance2
           using input order as the tie-break order
```

### 10.3 Target strategies

Use a small interface only here, because target computation is the intended variation point.

```text
TargetStrategy
  TargetFor(state, ghost) TilePos
```

Implementations:

| Strategy | Used by | Rule |
|---|---|---|
| `DirectPlayerTarget` | Blinky chase | Player current tile. |
| `FourAheadTarget` | Pinky chase | Four tiles ahead; preserve up-plus-left offset quirk. |
| `VectorTarget` | Inky chase | Double vector from Blinky through the two-ahead offset tile; preserve up-plus-left quirk. |
| `DistanceSwitchTarget` | Clyde chase | Player if distance is at least eight tiles, else Clyde scatter target. |
| `ScatterTarget` | All scatter | Fixed per-ghost unreachable corner target. |
| `HouseReturnTarget` | Eaten ghosts | Fixed tile leading to the ghost house entrance. |

Avoid strategy objects for every other behavior; direct functions or methods are clearer for stable gameplay logic.

### 10.4 Frightened routing

Frightened movement must use its own selector:

```text
FrightenedDirectionSelector
  ChooseDirection(ghost, decisionTile, random, maze) Direction
```

It selects a random starting direction, then rotates clockwise through candidates until it finds a legal non-reverse movement direction.

The RNG must be seeded on:

- Level start.
- Life reset after a player death.

For testing, the initial seed must be injectable through a simulation configuration object.

---

## 11. Mode Scheduler and Forced Reversals

### 11.1 Shared scheduler

`ModeSchedulerState` is global to the board. It owns an ordered list of mode segments from level configuration.

```text
ModeSegment
  BaseMode       Scatter | Chase
  DurationTicks  int; use a sentinel for indefinite

ModeSchedulerState
  SegmentIndex
  RemainingTicks
  Suspended      bool
```

All non-eaten ghosts derive their base behavior from this shared scheduler. Individual ghosts do not own independent chase/scatter clocks.

### 11.2 Transition processing

On a chase/scatter transition:

- Change the global base mode.
- Set `PendingReverse = true` for every eligible ghost.
- Mark ghosts currently inside the house as having seen a mode change.
- Do not reverse direction immediately.

On energizer consumption with an active fright duration:

- Suspend the base-mode scheduler.
- Put eligible normal ghosts in frightened mode.
- Set pending reversal for affected ghosts.
- Reset the ghost-eaten score chain.

When frightened ends:

- Restore each surviving ghost to the current base mode.
- Resume the scheduler.
- Do not force a reversal.

---

## 12. Ghost House Controller

This subsystem should be isolated because it combines release precedence, post-death counters, an anti-stall timer, and special exit direction behavior.

### 12.1 State

```text
GhostHouseState
  ReleasePolicy          LevelStartCounters | PostDeathGlobalCounter
  PersonalDotCounts      map[GhostIdentity]int
  GlobalDotCount         int
  GlobalCounterActive    bool
  NoDotTicks             int
  WaitingOrder           [Pinky, Inky, Clyde]
```

### 12.2 Responsibilities

- Register dots and energizers consumed by the player.
- Determine the currently preferred waiting ghost.
- Release by level-start personal counter policy.
- Release by post-death global checkpoint policy.
- Release via no-dot-eaten timeout.
- Track mode changes while a ghost is in the house.
- Choose left or right exit direction when a ghost departs.
- Revive eaten ghosts after they return to the house.

### 12.3 Priority recommendation

Apply release rules in this order per tick:

1. Complete physical return/revival operations.
2. Honor a release already scheduled by a previous tick.
3. Evaluate dot-counter checkpoints caused by newly consumed collectibles.
4. Evaluate the anti-stall timeout.

This keeps cause and effect predictable and makes test cases straightforward.

---

## 13. Collectibles, Scoring, and Fruit

### 13.1 Collectible storage

At runtime, collectibles should not be represented as hundreds of heap-allocated objects.

Recommended structure:

```text
CollectibleLayer
  Items []CollectibleType  // one entry per maze tile
  RemainingDots int
  DotsEatenThisLevel int
```

`CollectibleType` is a compact enum: `None`, `Dot`, `Energizer`.

This gives O(1) lookup/removal from a player tile and a direct remaining-dot count for Elroy and level completion.

### 13.2 Score state

```text
ScoreState
  CurrentScore             int
  HighScore                int
  GhostEatChainIndex       int
  FruitHistory             []FruitKind
```

Ghost chain values are read from a static array:

```text
[200, 400, 800, 1600]
```

On an energizer activation, reset the index to zero. Each eaten frightened ghost reads the current entry and increments the index, capped at the final value or constrained to the maximum legal number of ghosts.

### 13.3 Fruit state

```text
FruitState
  Visible                 bool
  Kind                    FruitKind
  Value                   int
  RemainingTicks          int
  SpawnThresholdsHit      [2]bool
```

Spawn at the first observed crossing of 70 and 170 dots eaten. Choose its lifespan from a bounded 9–10 second range with a dedicated deterministic RNG draw.

---

## 14. Level Configuration

### 14.1 Data-driven configuration

The game should load level definitions from JSON or a Go-embedded data file. JSON is recommended initially because balance tables remain easy to review and update without recompiling gameplay source.

```text
LevelRules
  PlayerSpeed
  GhostNormalSpeed
  GhostFrightenedSpeed
  GhostTunnelSpeed
  GhostEatenSpeed
  ElroyStage1RemainingDots
  ElroyStage2RemainingDots
  ElroyStage1Speed
  ElroyStage2Speed
  FrightenedDurationTicks
  FrightenedFlashCount
  GhostReleaseLimits
  NoDotReleaseTimeoutTicks
  ModeSchedule
  FruitKind
  FruitValue
```

### 14.2 Fallback rule for high levels

A resolver should return the explicitly configured level row when it exists and otherwise return the `21+` steady-state row.

```text
ResolveLevelRules(levelNumber):
    if exact row exists: return row
    return highest configured steady-state row
```

Keep this behavior inside `LevelRulesRepository`; no simulation system should contain level-number conditional chains.

---

## 15. Rendering and Ebiten Integration

### 15.1 Ebiten application type

The application package provides a type implementing Ebiten’s game interface. Its responsibilities are deliberately narrow:

- Poll input through `input.Adapter`.
- Advance the scene manager and simulation exactly once per Ebiten update.
- Draw the current scene using the renderer.
- Report the layout size.

Recommended logical game size: **224 × 288**, matching the canonical maze coordinate system. Use nearest-neighbor integer scaling or a configurable viewport scaler for modern window sizes.

### 15.2 Renderer architecture

```text
Renderer
  DrawScene(screen, scene, state)

WorldRenderer
  DrawMaze(screen, maze)
  DrawCollectibles(screen, state)
  DrawActors(screen, state)
  DrawFruit(screen, state)

HUDRenderer
  DrawScore(screen, state)
  DrawLives(screen, state)
  DrawLevelHistory(screen, state)
```

The renderer receives immutable or conventionally read-only snapshots. It must never call simulation methods that mutate state.

### 15.3 Animation

Animation frame selection is presentation logic, derived from simulation state:

- Player mouth frame derives from movement direction and animation tick.
- Ghost sprite derives from identity, mode, direction, and flash phase.
- Eaten ghost renders as eyes using direction.
- Fruit score popup derives from a score event with a short presentation-only lifetime.

Gameplay must not wait for animation completion except at explicit scene transitions such as death and level-complete sequences.

---

## 16. Input Design

Convert Ebiten keyboard input to a game-neutral frame value:

```text
InputFrame
  RequestedDirection     Direction
  StartPressed           bool
  PausePressed           bool
```

The simulation stores only the latest requested direction. It attempts that direction when legal under player cornering and maze rules.

This supports keyboard now and controller/remapping later without changing simulation code.

---

## 17. Event Model

Simulation events decouple rules from rendering and audio.

```text
GameEvent
  DotEaten
  EnergizerEaten
  FruitSpawned
  FruitEaten
  GhostEaten
  GhostReleased
  PlayerCaught
  ModeChanged
  LevelCleared
  ExtraLifeAwarded
```

Each event should include the minimum contextual data needed by consumers, such as tile, ghost identity, score value, and tick.

Events are ephemeral: generate them during `AdvanceTick`, consume them in the application layer, then clear the list before the next tick.

---

## 18. Testing Strategy

### 18.1 Unit tests

Focus on deterministic, rule-level units:

- Direction order and opposite-direction rules.
- Portal resolution.
- Candidate filtering in the ghost path planner.
- Tie-breaking order.
- Pinky and Inky up-direction offset quirks.
- Clyde’s eight-tile target threshold.
- Elroy threshold activation and suspension.
- Frightened selector rotation from blocked choices.
- Dot counters, global post-death counter, and anti-stall timer.
- Mode scheduler pause/resume behavior.

### 18.2 Scenario tests

Use short scripted sequences that initialize a precise board state, execute a number of ticks, and assert outcomes.

Examples:

| Scenario | Expected assertion |
|---|---|
| Mode transition while ghost is in corridor | Ghost reverses only at its next tile boundary. |
| Two equal-distance exits | Planner chooses Up, then Left, then Down, then Right. |
| Player and ghost swap adjacent tiles in one tick | No collision is registered. |
| Dot count reaches 70 | Fruit becomes visible exactly once. |
| Energizer is consumed | Scheduler pauses and eligible ghosts become frightened. |
| Frightened timer ends | Base mode resumes without a new reversal. |
| Post-death dot count reaches 7 | Pinky exits the house. |

### 18.3 Replay tests

Add optional input replay support early:

```text
Replay
  InitialSeed
  InitialLevel
  []InputFrame
```

A replay test hashes or serializes selected state after each tick and compares it with a known-good baseline. This is highly effective at detecting unintentional rule drift after refactoring.

---

## 19. Delivery Plan

### Milestone 1 — Deterministic maze movement

- Maze loading and validation.
- Player movement, wall collision, tunnel wrapping.
- Basic fixed-step simulation.
- Renderer for walls, dots, player, and debug tile overlays.
- Headless movement tests.

### Milestone 2 — Core ghosts

- Ghost movement and look-ahead decision scheduling.
- Shared path planner and scatter targets.
- Blinky, Pinky, Inky, Clyde target strategies.
- Tie-breaking and no-up-turn zones.
- Debug overlay for target tiles and planned direction.

### Milestone 3 — Modes and collisions

- Shared scatter/chase scheduler.
- Forced reversal semantics.
- Frightened mode with deterministic random routing.
- Same-tile collision, ghost-eaten return behavior.
- Death and reset sequence.

### Milestone 4 — House, scoring, and progression

- Personal and global release counters.
- Anti-stall release timer.
- Elroy behavior.
- Fruit, score chain, lives, level completion.
- Data-driven level table.

### Milestone 5 — Presentation polish

- Animated sprites, audio, HUD, level history strip.
- Menu/attract flow and pause.
- Accessibility options: scale, fullscreen, key remapping, reduced visual flashing.
- Regression replay suite.

---

## 20. Technical Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Mixing tile and pixel logic | AI routes, collisions, and turning become inconsistent. | Make `TilePos` and fixed-point `PixelPos` separate domain types; centralize conversions. |
| Variable timestep movement | Frame-counted mechanics diverge across machines. | Run all rules in fixed 60 Hz ticks. |
| Immediate mode reversals | Ghost behavior appears incorrect. | Use a per-ghost pending reversal flag consumed at tile entry. |
| Duplicate ghost AI implementations | Subtle behavior drift between ghosts. | One path planner; separate target strategies only. |
| Game rules inside renderer | Testing becomes difficult and bugs become timing-dependent. | Keep `sim` independent from Ebiten. |
| Floating-point accumulation | Rare tile-boundary and collision failures. | Use fixed-point integer positions. |
| Under-specified maze content | Accurate rules cannot compensate for incorrect topology. | Validate the maze with topology tests and visual debug tools. |

---

## 21. Recommended Debug Features

Enable through a development configuration flag only:

- Tile grid and coordinate labels.
- Current actor tile and sub-tile center marker.
- Ghost target tile marker in ghost color.
- Ghost planned direction and pending reversal marker.
- Current base mode, frightened timer, and mode segment timer.
- Ghost-house counters and no-dot timer.
- Dot count and Elroy stage.
- Input replay recording/playback.
- Single-tick stepping and simulation pause.

These features reduce the cost of validating the subtle rules that define faithful gameplay.

---

## 22. Final Recommendation

Build the implementation around a **headless deterministic simulation**, then attach Ebiten as a presentation and input shell. The most important design choices are:

1. Fixed-point actor positions plus separate tile occupancy.
2. A fixed 60 Hz tick pipeline with explicit update ordering.
3. One shared ghost path planner with isolated target strategies.
4. A shared mode scheduler plus per-ghost transition state.
5. A dedicated ghost-house controller.
6. Data-driven maze and level rules.
7. Scenario and replay tests for every subtle behavioral rule.

This structure provides an implementation that is faithful to the supplied design dossier while remaining maintainable, testable, and practical in idiomatic Go with Ebiten.
