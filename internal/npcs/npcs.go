package npcs

import "tektmud/internal/logger"

type NpcId int         //Represents the "reference" id of the NPC
type NpcInstanceId int //Represents the Npc in the world. i.e "SimpleName#InstanceId will show when doing advanced looking in a room."

var (
	nextNpcInstanceId uint64 = 3729       //Pick a fun number
	allNpcNames              = []string{} //Mostly used to validate player names at creation.
	npcInstances             = map[int]*NPC{}
	npcBlueprints            = map[NpcId]*NPC{} //holds a reference for all mobs we might be creating.
)

type NPC struct {
	Id          NpcId         `yaml:"id"`
	InstanceId  NpcInstanceId `yaml:"-"`
	AreaId      string        `yaml:"area,omitempty"`
	DefaultRoom string        `yaml:"-"`
	IsHostile   bool          `yaml:"is_hostile"`
	TetherMax   int           `yaml:"tether_max"` //How far can they wander from their default room.
	BuffIds     []int         `yaml:"buff_ids"`

	//Fields more related to impact for the player
	Level      int `yaml:"level"`
	XpAddMulti int `yaml:"xp_add_multi"` //Expressed as a %, where 10 = 110% xp value.
	HpBase     int `yaml:"hp_base"`      //Their Base HP which can be impacted by their level
	ManaBase   int `yaml:"mana_base"`    //Their base MP which can be impacted by their level
	Force      int `yaml:"force"`        //All mobs only use force to simplify my life for now

	angryAt map[uint64]struct{} //Any players this npc has attacked (or been attacked by) since spawning.
}

func NewNPCById(npcId NpcId, defaultRoom string, level ...int) *NPC {
	var actualLevel int = 0
	if len(level) > 0 {
		actualLevel = level[0]
	}

	if npc, exists := npcBlueprints[npcId]; exists {
		nextNpcInstanceId++
		n := *npc //Make a copy of the blueprint

		n.DefaultRoom = defaultRoom
		n.InstanceId = NpcInstanceId(nextNpcInstanceId)
		n.Level = max(actualLevel, n.Level)

		npcInstances[int(n.InstanceId)] = &n
		return npcInstances[int(n.InstanceId)]
	}

	logger.Warn("Attempted construction of npc that was unknown", "npcId", npcId, "room", defaultRoom)
	return nil
}

func GetAllNpcNames() []string {
	//return a copy just so we can do whatever with it
	return append([]string{}, allNpcNames...)
}

func (npc *NPC) IsAngryAt(entityId uint64) bool {
	if npc.angryAt == nil {
		return false
	}

	_, exists := npc.angryAt[entityId]
	return exists
}

func (npc *NPC) AttackedBy(entityId uint64) {
	if npc.angryAt == nil {
		npc.angryAt = map[uint64]struct{}{}
	}

	npc.angryAt[entityId] = struct{}{}
}
