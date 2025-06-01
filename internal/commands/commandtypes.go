package commands

type Input struct {
	UserId uint64
	Text   string
}

// Command interface
func (i Input) Name() string { return `Input` }

type Message struct {
	UserId          uint64   // Target
	ExcludedUserIds []uint64 // When used in rooms etc who not to show. ie.e room entry messsages
	RoomKey         string   // areaId:roomId
	Text            string
	IsCommunication bool // Is this affected by deafness ? say/shout/zone chat etc
}

// Command interface
func (m Message) Name() string { return `Message` }

type DisplayRoom struct {
	UserId  uint64 //Target
	RoomKey string
}

// Command interface
func (m DisplayRoom) Name() string { return `DisplayRoom` }

type PlayerQuit struct {
	UserId uint64
}

// Command interface
func (pq PlayerQuit) Name() string { return `PlayerQuit` }
