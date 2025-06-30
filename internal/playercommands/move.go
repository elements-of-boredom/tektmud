package playercommands

import (
	"fmt"
	"tektmud/internal/character"
	"tektmud/internal/players"
	"tektmud/internal/rooms"
)

func Move(args string, player *players.PlayerRecord, room *rooms.Room) (bool, error) {

	if len(args) == 0 {
		return false, fmt.Errorf("received a move command with no direction user:%v, room:%v", player.Id, room.Id)
	}
	//Check movement balance
	if !player.Char.Balance.HasBalance(character.MovementBalance) {
		player.SendText("You must wait before moving again.")
		return true, nil
	}

	if exit := room.FindExit(args); exit != nil {
		areaId, roomId := player.Char.GetLocation()

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

		//Before we send them to the room we need to attempt to Setup()
		//this ensures everything in the room is there before entering.

		dest.Setup()

		if err := rooms.MoveToRoom(player.Char, room, dest); err != nil {
			player.SendText("Unable to move into that room.")
		} else {
			player.Char.Balance.UseBalance(character.MovementBalance)
			//Notify everyone in the current room they left.
			room.SendText(
				fmt.Sprintf("%s leaves to the %s", player.Char.Name, string(exit.Direction)),
				player.Id)

			//Notify everyone in the new room they are entering
			dest.SendText(
				fmt.Sprintf("%s enters from the %s ", player.Char.Name, enterFromDirection),
				player.Id)

			dest.ShowRoom(player.Id)
		}

		return true, nil
	}

	//exit was nil. so they are attempting to go somewhere thats not valid for the room.
	//This can happen by typing "n" in a room w/ no north exit, OR if the user is blinded we want them
	//to have to guess.
	player.SendText("You cannot go in that direction.\n")

	//We handled this command (even though it failed), send back true
	return true, nil
}
