# Mission Rotation

## Purpose

The `class Missions {}` block in `server.cfg` defines an automatic mission cycle.
Without an admin, the server auto-selects missions when at least one player connects.
After a mission ends with players still on the server, it advances to the next entry.

ZEUS exposes this via the `config.missions` YAML key, which generates the `class Missions`
block inside `server.cfg`.

## Class Structure

```cpp
class Missions
{
    class ENTRY_NAME          // arbitrary identifier (letters/numbers/underscore, no spaces)
    {
        template = "MISSION.TERRAIN";
        difficulty = "regular";
        class Params           // optional: override mission parameters
        {
            PARAM_NAME = VALUE;
        };
    };
};
```

> **Warning:** The keyword `class` **must** be lowercase. A capitalized `Class` causes a
> parsing error.

## Template Naming

Format: `missionName.terrainName`

Three sources:

1. **Mission PBO** in MPMissions folder: `MyMission.MyTerrain.pbo`
   → template = `"MyMission.MyTerrain"`

2. **Mission folder** in MPMissions folder: `MyMission.MyTerrain/mission.sqm`
   → template = `"MyMission.MyTerrain"`

3. **Addon mission** loaded via mod: use the class name from `CfgMissions/MPMissions`,
   **not** the folder name. The class name and folder name are often different.

Example of an addon mission config:
```cpp
class CfgMissions {
    class MPMissions {
        class EscapeFromMalden   // class name — use this in template
        {
            directory = "A3\Missions_F_Patrol\MPScenarios\MP_EscapeFromMalden.Malden";
        };
    };
};
```
Template: `"EscapeFromMalden.Malden"` — not the directory folder name.

## Difficulty Levels

Standard values:

| Value | Description |
|-------|-------------|
| `"Recruit"` | Easy AI, assistance enabled |
| `"Regular"` | Default settings |
| `"Veteran"` | Harder AI, limited HUD |
| `"Custom"` | Reads from server's `.Arma3Profile` file |

Some mods and CDLCs add custom difficulty classes. Determine the class name using the
Arma 3 config viewer (`CfgDifficultyPresets`).

> **Note:** If difficulty is set in the mission cycle entry, it overrides `forcedDifficulty`
> from the global server config.

## Mission Parameters Override

Each mission in the cycle can override default mission parameters:

```cpp
class Missions
{
    class CombatPatrol01
    {
        template = "MP_CombatPatrol_01.Altis";
        difficulty = "veteran";
        class Params
        {
            BIS_CP_reinforcements = 2;  // default 0
            BIS_CP_tickets = 5;         // default 20
        };
    };
};
```

To find available parameters: extract the mission PBO and read `Description.ext`.

> **Warning:** `-autoInit` startup parameter breaks mission parameters — only default
> values are returned when autoInit is active.

## MP Campaign Collections

Missions grouped into campaigns (e.g. Apex Protocol) can be included as a whole:

```cpp
class Missions
{
    class Apex {};   // includes all missions in the Apex campaign class
};
```

Individual mission difficulty can be overridden within the campaign block:

```cpp
class Missions
{
    class Apex
    {
        class EXP_m01
        {
            difficulty = "veteran";
        };
    };
};
```

## Full Example

```cpp
class Missions
{
    class Mission01
    {
        template = "MP_Marksmen_01.Altis";
        difficulty = "recruit";
        class Params
        {
            RespawnDelay = 15;
        };
    };

    class Mission02
    {
        template = "EscapeFromMalden.Malden";
        difficulty = "regular";
    };

    class Mission03
    {
        template = "EXP_m01.Tanoa";    // Apex Protocol
        difficulty = "custom";
        class Params {};
    };
};
```

## Key Settings That Affect Mission Rotation

From `server.cfg`:
- `autoSelectMission` — required for automatic progression
- `randomMissionOrder` — shuffle instead of sequential
- `persistent` — keep mission running when server is empty
- `missionsToServerRestart`, `missionsToShutdown` — automatic recycling
- `votingTimeOut`, `roleTimeOut`, `briefingTimeOut`, `debriefingTimeOut` — phase time limits

From startup parameters:
- `-loadMissionToMemory` — cache mission in RAM for faster client loading
- `-autoInit` — initialize mission immediately (breaks mission parameters)

## Related

- [server.cfg Reference](server_cfg.md)
- [Startup Parameters](../protocols/startup_params.md)
