package usercommands

import (
	"fmt"
	"tektmud/internal/character"
	"tektmud/internal/rooms"
	"tektmud/internal/users"
)

func Move(args string, user *users.UserRecord, room *rooms.Room) (bool, error) {

	if len(args) == 0 {
		return false, fmt.Errorf("received a move command with no direction user:%v, room:%v", user.Id, room.Id)
	}
	//Check movement balance
	if !user.Char.Balance.HasBalance(character.MovementBalance) {
		user.SendText("You must wait before moving again.")
		return true, nil
	}

	if exit := room.FindExit(args); exit != nil {
		areaId, roomId := user.Char.GetLocation()

		//Parse the target destination
		destAreaId := areaId
		destRoomId := exit.Destination

		if len(exit.Description) > 0 && exit.Destination != roomId {
			parts := rooms.SplitDestination(exit.Destination)
			if len(parts) == 2 {
				destAreaId = parts[0]
				destRoomId = parts[1]
			}
		}

		//Validate the destination exists
		dest := rooms.LoadRoom(destAreaId, destRoomId)
		if dest == nil {
			return false, fmt.Errorf("found the exit but not the room, areaId:%s, roomId: %s, destAreaId:%s, destRoomId:%s", room.AreaId, room.Id, destAreaId, destRoomId)
		}

		enterFromDirection := dest.FindExitTo(areaId, roomId)
		if len(enterFromDirection) < 1 {
			enterFromDirection = "ether"
		}

		if err := rooms.MoveToRoom(user.Char, room, dest); err != nil {
			user.SendText("Unable to move into that room.")
		} else {
			user.Char.Balance.UseBalance(character.MovementBalance)
			//Notify everyone in the current room they left.
			room.SendText(
				fmt.Sprintf("%s leaves to the %s", user.Char.Name, string(exit.Direction)),
				user.Id)

			//Notify everyone in the new room they are entering
			dest.SendText(
				fmt.Sprintf("%s enters from the %s ", user.Char.Name, enterFromDirection),
				user.Id)

			dest.ShowRoom(user.Id)
		}

		return true, nil
	}

	//exit was nil. so they are attempting to go somewhere thats not valid for the room.
	//This can happen by typing "n" in a room w/ no north exit, OR if the user is blinded we want them
	//to have to guess.
	user.SendText("You cannot go in that direction.\n")

	//We handled this command (even though it failed), send back true
	return true, nil
}
