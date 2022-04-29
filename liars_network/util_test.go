package liars_network

import (
	"math"
	"testing"
)

func TestFindNetworkValue(t *testing.T) {
	test_elements_array_1 := []int32{100, 4, 7, 9, 100}
	if desired_element, exists := FindNetworkValue(test_elements_array_1, 2); !exists || desired_element != 100 {
		t.Errorf("%v should find 100 as the desired element", test_elements_array_1)
	}
	test_elements_array_2 := []int32{100, 4, 7, 7, 9, 100}
	if _, exists := FindNetworkValue(test_elements_array_2, 2); exists {
		t.Errorf("%v should not have any desired element because both 100 and 7 are of the same frequency", test_elements_array_2)
	}
	test_elements_array_3 := []int32{1, 1, 1, 2}
	if desired_element, exists := FindNetworkValue(test_elements_array_3, 3); !exists || desired_element != 1 {
		t.Errorf("%v should find 1 as the desired element", test_elements_array_3)
	}
}

func TestCheckStartOrExtendCommand(t *testing.T) {
	// cannot parse int/float64
	start_command_1 := "start --value v --max-value max --num-agents number --liar-ratio ratio"
	if CheckStartOrExtendCommand(start_command_1) != nil {
		t.Errorf(start_command_1, "should fail CheckStartOrExtendCommand() because the passed in args are parsable.")
	}
	start_command_2 := "start --hello"
	if CheckStartOrExtendCommand(start_command_2) != nil {
		t.Errorf(start_command_2, "should fail CheckStartOrExtendCommand() because the command is incomplete.")
	}
	start_command_3 := "start --value 10 --max-value 100 --num-agents 3 --liar-ratio 0.5"
	successful_map := CheckStartOrExtendCommand(start_command_3)
	if successful_map == nil {
		t.Errorf(start_command_3, "should succeed.")
	}
	if len(successful_map) != 4 {
		t.Errorf("successful_map should always be of size 4")
	}
	if successful_map["value"] != 10 {
		t.Errorf("Did not parse network_value correctly.")
	}
	if successful_map["max_value"] != 100 {
		t.Errorf("Did not parse max_value correctly.")
	}
	if successful_map["num_agents"] != 3 {
		t.Errorf("Did not parse num_agents correctly.")
	}
	if successful_map["liar_ratio"] != 0.5 {
		t.Errorf("Did not parse liar_ratio correctly.")
	}
	// wrong flag name passed in
	start_command_4 := "start --value 10 --name Goo --num-agents 3 --liar-ratio 0.5"
	if CheckStartOrExtendCommand(start_command_4) != nil {
		t.Errorf(start_command_4, "should fail CheckStartOrExtendCommand() because the flag passed in is unidentified")
	}
	// too many agents created
	start_command_5 := "start --value 10 --max-value 100 --num-agents 65536 --liar-ratio 0.5"
	if CheckStartOrExtendCommand(start_command_5) != nil {
		t.Errorf(start_command_5, "should fail CheckStartOrExtendCommand() because num_agents >= (2^16-1)")
	}
	// CheckStartOrExtendCommand() should still work with different ordering of key-pairs
	start_command_6 := "start --liar-ratio 0.5 --value 10 --max-value 100 --num-agents 3"
	successful_map = CheckStartOrExtendCommand(start_command_6)
	if successful_map == nil {
		t.Errorf(start_command_3, "should succeed.")
	}
	if len(successful_map) != 4 {
		t.Errorf("successful_map should always be of size 4")
	}
	if successful_map["value"] != 10 {
		t.Errorf("Did not parse network_value correctly.")
	}
	if successful_map["max_value"] != 100 {
		t.Errorf("Did not parse max_value correctly.")
	}
	if successful_map["num_agents"] != 3 {
		t.Errorf("Did not parse num_agents correctly.")
	}
	if successful_map["liar_ratio"] != 0.5 {
		t.Errorf("Did not parse liar_ratio correctly.")
	}
}

func TestCheckKillCommand(t *testing.T) {
	kill_command_1 := "kill --id -1"
	if _, valid := CheckKillCommand(kill_command_1); valid {
		t.Errorf(kill_command_1, "has an id out of range of [1, 65535].")
	}
	kill_command_2 := "kill --comany 123"
	if _, valid := CheckKillCommand(kill_command_2); valid {
		t.Errorf(kill_command_2, "has a unrecognized flag comany.")
	}
	kill_command_3 := "kill --id 123"
	if id, valid := CheckKillCommand(kill_command_3); !valid || id != 123 {
		t.Errorf(kill_command_3, "is a valid kill command.")
	}
}

func TestPlayExpertCommand(t *testing.T) {
	playexpert_command_1 := "playexpert --num-agents 3 --liar-ratio 0.5"
	successful_map := CheckPlayExpertCommand(playexpert_command_1, math.MaxInt64)
	if successful_map == nil {
		t.Errorf(playexpert_command_1, "should succeed.")
	}
	if len(successful_map) != 2 {
		t.Errorf("successful_map should always be of size 2.")
	}
	if successful_map["num_agents"] != 3 {
		t.Errorf("Did not parse num_agents correctly.")
	}
	if successful_map["liar_ratio"] != 0.5 {
		t.Errorf("Did not parse liar_ratio correctly.")
	}

	playexpert_command_2 := "playexpert --num-agents --liar-ratio 3 0.5"
	if CheckPlayExpertCommand(playexpert_command_2, math.MaxInt64) != nil {
		t.Errorf(playexpert_command_2, "should fail because a flag value is not followed immediately after a flag.")

	}

	playexpert_command_3 := "playexpert --network cosmos --liar-ratio 0.5"
	if CheckPlayExpertCommand(playexpert_command_3, math.MaxInt64) != nil {
		t.Errorf(playexpert_command_2, "should fail because --network is an unidentified flag.")
	}

	playexpert_command_4 := "playexpert --num-agents 3 --liar-ratio 0.5"
	if CheckPlayExpertCommand(playexpert_command_4, 1) != nil {
		t.Errorf(playexpert_command_2, "should fail because num_agents should be less than the number of running agents.")
	}
}
