# Space MUD Design Summary

## Theme
Space setting (similar to Ironsworn: Starforged)

## Stats (4 total)
- **Force** (physical power)
- **Reflex** (damage avoidance)
- **Acuity** (technical + magical damage)
- **Heart** (health, resilience, endurance)

## Races (7 total)

### Humans
**Description:** The most adaptable species in known space, humans have spread across countless worlds and adapted to diverse environments. Their balanced capabilities and shorter lifespans drive them to achieve more in less time. They serve as diplomats, traders, and generalists throughout the galaxy. Their flexibility makes them suitable for any class or playstyle.

**Stats:** Force 12, Reflex 12, Acuity 12, Insight 12, Heart 12
**Special:** Slightly reduced XP requirements for leveling

### Synthetics  
**Description:** These are human consciousness housed in artificial bodies, combining biological thought with mechanical precision. Created through advanced medical procedures that preserve the brain while replacing everything else, they excel in technical fields and dangerous environments where their durable frames provide advantages. Despite their artificial bodies, they maintain fully human personalities and desires.

**Stats:** Force 12, Reflex 11, Acuity 14, Insight 11, Heart 12
**Special:** Immune to biological hazards, require maintenance instead of food/sleep

### Umbrans
**Description:** A mysterious alien species that exists partially outside normal reality, Umbrans appear as shifting, shadowy humanoid figures. Their natural ability to phase between dimensions makes them master infiltrators and scouts. They communicate through subtle mental emanations and are rarely seen unless they choose to be. Their otherworldly nature makes them excellent at stealth and espionage operations.

**Stats:** Force 10, Reflex 15, Acuity 12, Insight 13, Heart 10
**Special:** Natural stealth bonuses, phasing abilities

### Verdani
**Description:** Plant-based humanoids with bark-like skin and photosynthetic capabilities, the Verdani are living bridges between technology and nature. They can supplement their nutrition through light absorption and have natural regenerative abilities. Their deep understanding of biological systems makes them exceptional at life sciences and terraforming. They move deliberately but possess incredible resilience and patience.

**Stats:** Force 11, Reflex 10, Acuity 12, Insight 12, Heart 15
**Special:** Reduced food requirements, natural regeneration in light, biological sciences bonuses

### Voidborn
**Description:** Descendants of humans who spent generations traveling between stars, the Voidborn have adapted to life in deep space. Their pale, elongated forms and enlarged eyes reflect their comfort in low-gravity, low-light environments. They possess an intuitive understanding of stellar navigation and cosmic phenomena. Having never known planetary life, they feel most at home in the endless void between worlds.

**Stats:** Force 10, Reflex 12, Acuity 14, Insight 14, Heart 10
**Special:** Natural navigation abilities, zero-G movement bonuses, stellar phenomenon detection

### Stoneheart
**Description:** Massive, rocky humanoids with silicon-based physiology and incredible durability, the Stoneheart are living mountains. Their crystalline formations can interface directly with technological systems, making them natural engineers despite their imposing appearance. They speak slowly and deliberately, thinking in geological timescales. Their incredible strength and resilience make them nearly unstoppable once they commit to action.

**Stats:** Force 15, Reflex 8, Acuity 11, Insight 10, Heart 16
**Special:** Natural armor, technology interfacing, resistance to environmental hazards

### Corvans
**Description:** Highly intelligent avian humanoids descended from enhanced corvids, the Corvans combine fierce intellect with incredible dexterity. Their feathered bodies retain vestigial wings and their eyes miss nothing, making them exceptional scouts and technicians. They exhibit the problem-solving intelligence of their corvid ancestors amplified by genetic enhancement. Their quick wit and nimble fingers make them masters of complex tasks requiring precision and creativity.

**Stats:** Force 9, Reflex 16, Acuity 14, Insight 12, Heart 9
**Special:** Enhanced problem-solving, superior manual dexterity, keen eyesight

## Damage Types (8 total)
1. **Fire**
2. **Electrical**
3. **Cold**
4. **Blunt**
5. **Slashing**
6. **Poison**
7. **Radiation** (includes psionic effects)
8. **Sonic**

## Afflictions (17 total)
1. **Sluggish** - Increased balance timers (makes everything take longer)
2. **Disoriented** - Random chance to move in wrong direction
3. **Taunted** - Cannot leave room without overcoming the taunt
4. **Pacified** - Cannot initiate attacks (but can defend)
5. **Suppressed** - Cannot use special abilities, only basic attacks
6. **Overloaded** - Risk taking damage when dealing damage
7. **Shocked** - Cannot take any actions for X rounds
8. **Drained** - Mana/energy regeneration stopped
9. **Bleeding** - Continuous health loss over time
10. **Poisoned** - Lethal damage after a set period if not cured
11. **Radiation Sickness** - Lethal damage after X time if not cured
12. **Broken Limb** - Specific penalties based on limb
13. **Blinded** - Cannot see room descriptions, locations, targeting info
14. **Confused** - Actions target random people/objects
15. **Terrified** - Forced to flee or reduced effectiveness near specific enemies
16. **Disarmed** - Cannot use weapon-based attacks until re-equipped
17. **Burning** - Continuous fire damage over time

## Class Damage Specializations
- **Vanguard** - Blunt, Slashing
- **Mentalist** - Radiation, Electrical
- **Fabricator** - Poison, Electrical
- **Animist** - Radiation
- **Warden** - Blunt, Cold
- **Symbiont** - Varies by creature bond
- **Distortionist** - Blunt, Radiation
- **Harmonist** - Sonic
- **Augur** - Poison
- **Scavenger** - Fire, Electrical

## Classes (10 regular + 1 special)

### 1. Vanguard (Force-based melee + tactical leadership)
**Primary Skillset - Warfare:**
- Basic Attack (standard melee)
- Intimidating Strike (demoralizes enemies, reducing damage output) [CAPSTONE]
- 8 combat skills [PLACEHOLDER]

**Support Skillset - Armament (10 skills):**
*Weapon Modifications:*
- Thermal Edge (convert to Fire damage)
- Shock Treatment (convert to Electrical damage)
- Cryo Coating (convert to Cold damage)
- Swift Balance (increase attack speed, trade damage for speed)
- Weapon Soul (manifest/create ideal bonded weapon) [CAPSTONE]

*Armor Enhancements:*
- Life Support Boost (enhanced health/stamina recovery)
- Adaptive Plating (resistance to last damage type, vulnerability to opposite)
- Adrenaline Surge (reduced damage window, then increased damage)
- Power Cycling (enhanced mana/energy recovery)
- Fortress Protocol (damage reduction, mobility penalty)

**Support Skillset - Tactics (10 skills):**
*Affliction Abilities:*
- Challenge (inflicts Taunted)
- Intimidating Presence (inflicts Confused)
- Overwhelming Force (inflicts Sluggish)
- Crippling Strike (inflicts Broken Limb)
- Suppressive Assault (inflicts Suppressed)

*Room-Wide Shout Abilities:*
- Rally Cry (cures Mental/Morale afflictions for all allies in room)
- Battle Command (cures Movement/Positioning afflictions for all allies in room)

*Tactical Support:*
- Formation Fighting (positioning bonuses when near another Vanguard)
- Battlefield Awareness (detect hidden enemies, assess threats)
- War Cry (inflicts Taunted, Confused, and Drained on all enemies in room) [CAPSTONE]

### 2. Mentalist (Acuity-based mental powers + energy manipulation)
**Primary Skillset - Psionics:**
- Mind Blast (basic attack)
- Mental Overload (feedback damage when enemy uses abilities) [CAPSTONE]
- 8 other psionic skills [PLACEHOLDER]

**Support Skillset - Radiation (10 skills):**
*Affliction Abilities:*
- Energy Drain (inflicts Drained)
- Blinding Flash (inflicts Blinded)
- Neural Static (inflicts Shocked)
- System Slowdown (inflicts Sluggish)
- Power Suppression (inflicts Suppressed)
- Radiation Burns (inflicts Bleeding)
- Mind Scramble (inflicts Confused)
- Fatal Exposure (inflicts Radiation Sickness)

*Utility:*
- X-Ray Vision (see through doors/walls using radiation)

*Capstone:*
- Dual Emission (cast two different radiation afflictions simultaneously) [CAPSTONE]

**Support Skillset - BioElectrics (10 skills):**
- Reflexive Charge (enhance ally's Reflex/reaction speed)
- Neural Boost (increases ally's mana regeneration)
- Power Surge (slight DPS boost at cost of minor mana drain, self-cast only)
- Energy Transfer (give your mana/energy to allies)
- Circuit Healing (create bio-electric healing device with N charges, tradeable, scales with skill level)
- Chemotherapy (purge 2 afflictions at cost of health and mana)
- Electrical Shield (reduce electrical, fire, radiation, sonic damage)
- Bioelectric Scan (analyze enemy weaknesses/resistances)
- Static Field (create area that provides minor HP boost to allies)
- Amplify (double ally's damage on next attack, heavy mana drain) [CAPSTONE]

### 3. Fabricator (Creates tech/bio agents for passive damage)
**Primary Skillset - Rigger:**
- Basic Attack (deploy basic attack drone)
- Swarm Burst (deploy multiple small attack drones) [CAPSTONE]
- 8 other drone-related skills [PLACEHOLDER]

**Support Skillset - Corruptor (10 skills):**
*Affliction Abilities (Nanobot-themed):*
- Viral Agent (inflicts Poisoned)
- Hemorrhagic Toxin (inflicts Bleeding)
- Nerve Agent (inflicts Sluggish)
- Energy Parasite (inflicts Drained)
- Hallucinogen (inflicts Confused)
- Muscle Relaxant (inflicts Suppressed)
- Paralytic Serum (inflicts Shocked)

*Utility Abilities:*
- Purge Protocol (strip buffs from enemy)
- Transfer Corruption (move random debuff from self to enemy)

*Capstone:*
- Nanobot Recall (retrieve all nanobots from target for damage based on number of afflictions) [CAPSTONE]

**Support Skillset - Construction (10 skills):**
- Life Support System (room-wide HP/mana regeneration for allies)
- Resource Extractor (create permanent items/materials in the world)
- Surveillance Bug (plant listening device in room, hidden)
- Barrier Wall (block doors/passages temporarily)
- Jet Boots (grant self flying capability)
- Emergency Breather (craft device for no-oxygen environments)
- [PLACEHOLDER - Permanent Item Creation]
- [PLACEHOLDER - Permanent Item Creation]
- [PLACEHOLDER - Permanent Item Creation]
- Portal Network (create/manage 1 pair of permanent portals for ally travel) [CAPSTONE]

### 4. Animist (Works with life spirit/soul energy)
**Primary Skillset - Soul Magic:**
- Basic Attack (standard damage)
- Life Drain (reduced damage but heals self for % of damage dealt) [CAPSTONE]
- 8 other soul-based skills [PLACEHOLDER]

**Support Skillset - Vitality (11 skills):**
- Poison Immunity (channeled buff with mana drain)
- Physical Resilience (reduce duration/effect of physical afflictions)
- Elemental Resilience (reduce duration/effect of elemental afflictions)
- Mental Resilience (reduce duration/effect of mental afflictions)
- Life Force (restore both health and mana simultaneously)
- Emergency Heal (automatic trigger when health drops low)
- Vital Restoration (raw heal using mana for balance)
- Life Anchor (create object that absorbs partial damage meant for Animist)
- Vitality Boost (10% HP bonus buff for self or allies)
- Life Sense (detect living creatures and see their health status)
- Purge (remove all afflictions from self, substantial 3x cooldown) [CAPSTONE]

**Support Skillset - Essence (10 skills):**
- Animate Living (make living creatures/NPCs/animals speak and communicate)
- Animate Object (give temporary life/speech to non-living objects)
- Spirit Shock (inflicts Shocked)
- Life Disruption (inflicts Drained)
- Bone Break (inflicts Broken Limb)
- Transfer Life (move health between targets)
- Sanctuary (prevents all combat actions in room - channeled, mana drain, broken by any other action)
- [PLACEHOLDER]
- [PLACEHOLDER]
- Soul Link (link soul with player/NPC to cast essence spells on them from anywhere in same zone) [CAPSTONE]

### 5. Warden (Protection + safe zones + threat detection)
**Primary Skillset - Aegis:**
- Basic Attack (Shield Bash)
- Retaliatory Strike (raise shield and reflect all damage back to attackers for N seconds) [CAPSTONE]
- 8 other aegis/shield skills [PLACEHOLDER]

**Support Skillset - Vigilance (10 skills):**
- Weapon Retention (avoid being disarmed)
- Clear Sight (avoid blindness or cure it rapidly)
- Survey (see into adjacent rooms - free version of universal skill)
- Enemy Analysis (see enemy weaknesses/resistances)
- Zone Watch (receive notifications of people entering/exiting current zone)
- Intercept (pull enemy from adjacent room)
- Heightened Reflexes (haste-like ability)
- Hidden Detection (reveal concealed enemies, traps, or objects in area)
- Combat Awareness (see all combatants' health/status at a glance)
- Danger Sense (enhanced awareness with chance to fully dodge attacks) [CAPSTONE]

**Support Skillset - Patrol (10 skills):**
- Emergency Response (teleport to ally under attack)
- Patrol Dash (move through multiple rooms in one direction, ignoring movement balance)
- Track (quickly identify location of tracked target within current zone)
- Secure Perimeter (plant Beacon that temporarily buffs self or ally with 10% HP)
- Quick Response (enhanced movement speed when responding to threats)
- Freezing Strike (inflicts Shocked)
- Bone Crusher (inflicts Broken Limb)
- Disarming Blow (inflicts Disarmed)
- Stunning Impact (inflicts Confused)
- Zone Teleport (immediately teleport to any target within the zone) [CAPSTONE]

### 6. Symbiont (Bonds with single powerful alien creature)
**Primary Skillset - Bonding:**
- Basic Attack (bonded creature performs attack)
- **Creature Bond Options:**
  - **Gelix** (blobbish plasma entity - Electrical damage): Amorphous plasma-like entity that can reshape and manipulate energy/electrical effects
  - **Nexari** (swarming collective - adaptive damage): Collective consciousness of tiny insectoid entities that share thoughts and coordinate perfectly
  - **Drakmor** (bear-bird hybrid - Blunt damage): Powerful creature with massive claws/strength like a bear but with keen aerial senses and limited flight capability
  - **Veilkin** (phase being - Radiation damage): Ethereal being that exists partially outside normal reality, can phase between dimensions
- Paired Assault (fusion attack - temporary merge with creature for enhanced attack) [CAPSTONE]
- 6 other bonding/creature management skills [PLACEHOLDER]

**Support Skillset - Corruption (11 skills):**
*Enemy Corruption:*
- Metabolic Slowdown (inflicts Sluggish)
- Parasitic Drain (inflicts Drained)
- Cellular Breakdown (inflicts Bleeding)
- Toxic Secretion (inflicts Poisoned)
- Sensory Rot (inflicts Blinded)
- Pheromone Terror (inflicts Terrified)
- Neural Overload (inflicts Overloaded)

*Self/Symbiote Support:*
- Fearless Bond (immunity to Terrified)
- Symbiotic Cleansing (symbiote automatically removes affliction from you, mana drain)
- Life Siphon (take life from symbiote to heal yourself, limited use, can kill symbiote)

*Capstone:*
- Chaos Corruption (apply 4 random afflictions to enemy and 3 random to yourself, no text explanation of which afflictions) [CAPSTONE]

**Support Skillset - Adaptation (10 skills):**
- Hardened Carapace (damage resistance for both Symbiont and creature)
- Enhanced Senses (see hidden things, traps, concealed objects)
- Adaptive Camouflage (invisible to regular sight)
- Integration (breathe in non-oxygen environments + prevent Blinded affliction)
- Metabolic Efficiency (reduce resource consumption by N%)
- Symbiotic Vigor (increase Heart and Force by 1 through creature bond)
- Symbiotic Mount (ride creature except Gelix, prevents Disoriented affliction)
- Remote Listening (send creature to adjacent room to report what's said)
- Enhanced Regeneration (accelerated healing when out of combat)
- Location Memory (creature takes you to any previously visited location - teleport with movement flavor) [CAPSTONE]

### 7. Distortionist (Exists across multiple dimensions/timelines - No direct attacks, wins through reflection)
**Primary Skillset - Compulsion:**
- Provoke (force enemy to attack you - basic "attack")
- Berserker Trigger (enemy enters rage state, attacks rapidly but recklessly) [CAPSTONE]
- 8 other compulsion/manipulation skills [PLACEHOLDER]

**Support Skillset - Distortion (10 skills):**
*Affliction Abilities (non-attack-stopping):*
- Feedback Loop (inflicts Overloaded)
- Hemorrhagic Warp (inflicts Bleeding)
- Spatial Drain (inflicts Drained)
- Perception Scramble (inflicts Confused)
- Dimensional Terror (inflicts Terrified)
- Sensory Distortion (inflicts Blinded)
- Phase Disruption (inflicts Shocked)
- Reality Sickness (inflicts Radiation Sickness)

*Utility:*
- Mirror Shield (increase damage reflection percentage)

*Capstone:*
- Reality Mirror (create clones of attacker that mirror their attacks back at the original attacker) [CAPSTONE]

**Support Skillset - Flux (10 skills):**
- Dimensional Escape (move when off balance, decent cooldown)
- Phase Walk (move through most closed doors)
- Temporal Shield (general damage reduction)
- Reality Anchor (remove mental afflictions: Sluggish, Disoriented, Taunted, Suppressed, Pacified, Confused)
- Reality Sense (see hidden things, detect dimensional threats/anomalies)
- Dimensional Vision (see whatever is in the room of a specific target)
- Reality Mend (cure broken limbs by distorting reality to restore them)
- Time Snapshot (save health/MP state, restore within 10 seconds OR flat heal if beyond window)
- [PLACEHOLDER]
- Void Transport (move ally and self to the Void - fast travel nexus) [CAPSTONE]

### 8. Harmonist (Sound/vibration manipulation)
**Primary Skillset - Harmonics:**
- Basic Attack (sonic-based attack)
- Frequency Burst (applies Shocked, Blinded and does decent damage) [CAPSTONE]
- 8 other harmonic skills [PLACEHOLDER]

**Support Skillset - Resonance (10 skills):**
*Room-Based Effects:*
- Healing Harmony (room-wide healing sound for all allies)
- Resonant Recovery (room-wide mana regeneration sound)
- Protective Frequency (room-wide damage reduction field)
- Harmonic Shield (room field that reduces Fire, Cold, Electrical, and Sonic damage)
- Sound Cancellation (prevent talking/sound-based abilities in room)
- Resonant Cleansing (room removes afflictions over time)
- Sound Trap (trap that announces in zone chat when triggered)

*Sound-Based Crafting:*
- Resonant Jewelry (create special jewelry that only Harmonist can wear)
- Universal Harmonics (create regular amulets that anyone can wear)
- Masterpiece Harmony (create ultimate harmonic amulet attuned to creation location, allows teleport back to that room) [CAPSTONE]

**Support Skillset - Vibrations (10 skills):**
*Sound-Based Afflictions:*
- Echo Confusion (inflicts Confused)
- Subsonic Terror (inflicts Terrified)
- Harmonic Disruption (inflicts Disoriented)
- Frequency Jam (inflicts Suppressed)
- Sonic Overload (inflicts Shocked)
- Resonant Drain (inflicts Drained)
- Sound Damage (inflicts Bleeding)

*Vibration Utilities:*
- Bone Shatter (inflicts Broken Limb)
- Ground Tremor (inflicts Disarmed)

*Capstone:*
- Sonic Amplification (room-wide effect that increases sonic damage taken by all enemies) [CAPSTONE]

### 9. Augur (Curse/debuff specialist)
**Primary Skillset - Hexcraft:**
- Basic Attack (damaging curse)
- Death Curse (instant death for anything under 30% life, requires at least 2 curses on target) [CAPSTONE]
- 8 other hexcraft skills [PLACEHOLDER]

**Support Skillset - Chronicle (11 skills):**
*Chronicle-Based Afflictions:*
- Chronicle of Pain (inflicts Bleeding)
- Chronicle of Weakness (inflicts Sluggish)
- Chronicle of Madness (inflicts Confused)
- Chronicle of Terror (inflicts Terrified)
- Chronicle of Blindness (inflicts Blinded)
- Chronicle of Silence (inflicts Suppressed)
- Chronicle of Exhaustion (inflicts Drained)

*Doom Writing Mechanics:*
- Fate Sealed (passive: all afflictions cast by Augur are harder to remove)
- Final Chapter (escalating doom over time - gets worse if not cured)
- Erase Entry (remove afflictions you've written)

*Capstone:*
- Destiny Written (write multiple afflictions simultaneously or apply devastating combined doom) [CAPSTONE]

**Support Skillset - Decree (10 skills):**
*Individual Buff Decrees:*
- Decree of Vitality (command improved health/endurance)
- Decree of Swiftness (command enhanced speed/reflexes)
- Decree of Strength (command increased physical power)

*Room Effect Decrees:*
- Decree of Lack (room-wide HP/MP degeneration)
- Decree of Restoration (room-wide HP/MP regeneration)
- Decree of Truth (prevent stealth and reveal hidden things in room)
- Decree of Binding (prevent movement in/out of room)
- Decree of Terror (room causes Terrified affliction)

*Decree Management:*
- Amendment (remove decree written in room)

*Capstone:*
- Decree of Ward (make self immune to one non-basic damage type: Fire/Cold/Electrical) [CAPSTONE]

### 10. Scavenger (Salvage, adaptation, resourcefulness)
**Primary Skillset - Contraption:**
- Basic Attack (Scrap Fling - reach into bag and throw random scrap item)
- Scrap Grenade (large damage to all enemies in room + inflicts Burning) [CAPSTONE]
- 8 other contraption skills [PLACEHOLDER]

**Support Skillset - Gadgetry (11 skills):**
*Harmful Contraptions (Afflictions):*
- Shock Bomb (applies Shocked)
- Poison Bomb (applies Poisoned)
- Shrapnel Bomb (applies Bleeding)
- Concussion Bomb (applies Broken Limb)
- Flash Bomb (applies Blinded)
- Glue Bomb (applies Sluggish)

*Improvised Medical Devices:*
- Jury-Rigged Medkit (makeshift healing device)
- Field Stimulant (grants haste effect)
- Makeshift Bandages (cure Bleeding and Broken Limb afflictions)
- Scrap Antidote (cure poison)

*Capstone:*
- Master Tinkerer (deploy tiny robot that slowly walks toward enemy; instant kill if not destroyed by basic attack within X seconds) [CAPSTONE]

**Support Skillset - Salvage (11 skills):**
*Economic/Trade:*
- Scrap Dealer (passive: increase sales to vendors by X%)
- [Unique Drop System] (Scavengers get special salvage materials others can't find)
- Salvage Map (craft map that leads back to creation point - portal using unique materials)

*Crafting:*
- Salvage Cache (create storage box/chest for items)
- Scrap Helmet (craft special helmet that grants damage reduction, requires N salvage materials)

*Detection/Information:*
- Keen Eye (detect hidden objects and secrets)
- Salvager's Intuition (see all combatants' information)

*Personal Enhancement:*
- Salvager's Endurance (grants increased Reflex and Heart stats)
- Makeshift Power (DPS boost ability)

*Utility:*
- Waste Not (resource cost reduction buff)

*Capstone:*
- Ultimate Repurpose (turn N unique currency into random consumables/buffs: HP/Mana potions, temporary resilience potions, etc.) [CAPSTONE]

### Special Unlockable Class
**11. Remnant Echo** (temporarily transforms into other classes with limited usage)

**Concept**: Player can temporarily become another class for a limited time with long cooldowns. Unlocked through quest chain starting with a Distortionist NPC who introduces the player to alternate timelines where they are different classes.

**Mechanics**:
- Transform into any unlocked class for 2-4 hours of active play time
- Cooldown of one game month (approximately one real-life week) between uses
- Cannot transform while in combat, requires preparation time
- Must unlock each class individually through quests/encounters

**Quest Progression**:
- Initial quest chain involves Distortionist NPC creating dimensional rifts
- Player encounters alternate timeline versions of themselves as different classes
- Defeating each alternate version unlocks that class as a "remnant echo"
- Each victory leaves behind fragments of that timeline's class abilities

**Restrictions**:
- Only have access to abilities the player has personally seen/experienced
- Limited time per transformation prevents exploitation
- Cannot combine abilities from multiple classes simultaneously

## Class Structure
- Each class has 1 **Primary** skillset (contains basic attack + thematic special attack)
- Each class has 2 **Support** skillsets
- Variable skills per skillset (starting baseline: 10 skills)
- All thematic special attacks mechanically balanced but flavored differently

## Core Game Systems

### Progression
- Separate "lesson" currency from experience for skill training
- Experience primarily from killing NPCs, some from mini-quests
- Players plan skill advancement at level milestones
- 100 levels total with accelerated early progression for retention

#### XP Progression Table (1-100)

**Levels 1-25 (Early Game - Fast Hook)**

| Level | XP Required | Cumulative XP | Level | XP Required | Cumulative XP |
|-------|-------------|---------------|-------|-------------|---------------|
| 1     | 250         | 250           | 14    | 2,000       | 13,900        |
| 2     | 300         | 550           | 15    | 2,250       | 16,150        |
| 3     | 400         | 950           | 16    | 2,500       | 18,650        |
| 4     | 500         | 1,450         | 17    | 2,750       | 21,400        |
| 5     | 600         | 2,050         | 18    | 3,000       | 24,400        |
| 6     | 750         | 2,800         | 19    | 3,250       | 27,650        |
| 7     | 850         | 3,650         | 20    | 3,500       | 31,150        |
| 8     | 950         | 4,600         | 21    | 3,750       | 34,900        |
| 9     | 1,100       | 5,700         | 22    | 4,200       | 39,100        |
| 10    | 1,250       | 6,950         | 23    | 4,500       | 43,600        |
| 11    | 1,500       | 8,450         | 24    | 4,750       | 48,350        |
| 12    | 1,650       | 10,100        | 25    | 5,000       | 53,350        |
| 13    | 1,800       | 11,900        |       |             |               |

**Levels 26-50 (Mid Game - Fast Progression)**
| Level | XP Required | Cumulative XP | Level | XP Required | Cumulative XP |
|-------|-------------|---------------|-------|-------------|---------------|
| 26    | 5,500       | 58,850        | 39    | 11,500      | 169,950       |
| 27    | 5,800       | 64,650        | 40    | 12,000      | 181,950       |
| 28    | 6,200       | 70,850        | 41    | 12,500      | 194,450       |
| 29    | 6,600       | 77,450        | 42    | 13,000      | 207,450       |
| 30    | 7,000       | 84,450        | 43    | 13,500      | 220,950       |
| 31    | 7,500       | 91,950        | 44    | 14,000      | 234,950       |
| 32    | 8,000       | 99,950        | 45    | 14,500      | 249,450       |
| 33    | 8,500       | 108,450       | 46    | 15,000      | 264,450       |
| 34    | 9,000       | 117,450       | 47    | 15,500      | 279,950       |
| 35    | 9,500       | 126,950       | 48    | 16,000      | 295,950       |
| 36    | 10,000      | 136,950       | 49    | 16,500      | 312,450       |
| 37    | 10,500      | 147,450       | 50    | 17,000      | 329,450       |
| 38    | 11,000      | 158,450       |       |             |               |

**Levels 51-75 (Transition - Slowing Progression)**
| Level | XP Required | Cumulative XP | Level | XP Required | Cumulative XP |
|-------|-------------|---------------|-------|-------------|---------------|
| 51    | 18,000      | 347,450       | 64    | 40,000      | 717,450       |
| 52    | 19,000      | 366,450       | 65    | 42,000      | 759,450       |
| 53    | 20,000      | 386,450       | 66    | 44,000      | 803,450       |
| 54    | 21,000      | 407,450       | 67    | 46,000      | 849,450       |
| 55    | 22,000      | 429,450       | 68    | 48,000      | 897,450       |
| 56    | 24,000      | 453,450       | 69    | 50,000      | 947,450       |
| 57    | 26,000      | 479,450       | 70    | 52,000      | 999,450       |
| 58    | 28,000      | 507,450       | 71    | 54,000      | 1,053,450     |
| 59    | 30,000      | 537,450       | 72    | 56,000      | 1,109,450     |
| 60    | 32,000      | 569,450       | 73    | 58,000      | 1,167,450     |
| 61    | 34,000      | 603,450       | 74    | 60,000      | 1,227,450     |
| 62    | 36,000      | 639,450       | 75    | 62,000      | 1,289,450     |
| 63    | 38,000      | 677,450       |       |             |               |

**Levels 76-94 (High Level - Significant Investment)**
| Level | XP Required | Cumulative XP | Level | XP Required | Cumulative XP |
|-------|-------------|---------------|-------|-------------|---------------|
| 76    | 64,000      | 1,353,450     | 86    | 84,000      | 2,103,450     |
| 77    | 66,000      | 1,419,450     | 87    | 86,000      | 2,189,450     |
| 78    | 68,000      | 1,487,450     | 88    | 88,000      | 2,277,450     |
| 79    | 70,000      | 1,557,450     | 89    | 90,000      | 2,367,450     |
| 80    | 72,000      | 1,629,450     | 90    | 92,000      | 2,459,450     |
| 81    | 74,000      | 1,703,450     | 91    | 94,000      | 2,553,450     |
| 82    | 76,000      | 1,779,450     | 92    | 96,000      | 2,649,450     |
| 83    | 78,000      | 1,857,450     | 93    | 98,000      | 2,747,450     |
| 84    | 80,000      | 1,937,450     | 94    | 100,000     | 2,847,450     |
| 85    | 82,000      | 2,019,450     |       |             |               |

**Levels 95-100 (End Game - Target Zone)**
| Level | XP Required | Cumulative XP |
|-------|-------------|---------------|
| 95    | 102,000     | 2,949,450     |
| 96    | 104,000     | 3,053,450     |
| 97    | 106,000     | 3,159,450     |
| 98    | 108,000     | 3,267,450     |
| 99    | 110,000     | 3,377,450     |
| 100   | 112,000     | 3,489,450     |

**Key Metrics:**
- Levels 1-50: 329,450 total XP (fast retention)
- Level 99 total: 3,377,450 XP
- Level 100 total: 3,489,450 XP
- End game target: 1 hour of play = 25-50% of a level
- Remnant Echo: Introduced at level 95, unlocked at level 100

### Economy & Equipment
- Limited equipment slots (1-3 total)
- **Credits** drive all cost-based activities, drop from mobs (limited amounts), primarily from quest turn-ins
- **Stellar Dust** is an earnable currency obtained via quests and other special means
- Equipment focus on character progression rather than gear accumulation

### Space Travel
- Mix of instant travel and active navigation for dangerous routes
- Skill checks and consequences during hazardous journeys
- Character skills affect travel outcomes

### Death/Resurrection
- Clone backup system (players choose when to update)
- Emergency beacon system for rescue timing/location
- Multiple death/resurrection modes (standard, clone backup, hardcore permadeath)

### General Skillsets
- 1-2 non-class skillsets available to all players
- Examples: Survival (fishing, death notifications, endurance bonuses)

## Design Principles
- Balance simplicity with depth
- All classes relatively balanced (not rigid tank/DPS/healer roles)
- Character-focused rather than equipment-focused progression
- Space theme integrated into mechanics, not just flavor
- Afflictions create tactical depth in PvP beyond pure damage