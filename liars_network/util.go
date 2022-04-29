package liars_network

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Identifies the element of the desired frequency from the input slice. This number
// needs to unique. If there are multiple elements of the same frequency in the slice,
// then it fails to find such element.
func FindNetworkValue(elements []int32, desired_frequency int) (int32, bool) {
	element_to_frequency_map := map[int32]int{}
	var network_value int32
	var same_frequency_element_num int32
	for _, element := range elements {
		element_to_frequency_map[element]++
	}
	for element, frequency := range element_to_frequency_map {
		if frequency == desired_frequency {
			network_value = element
			same_frequency_element_num++
		}
	}
	if same_frequency_element_num != 1 {
		return 0, false
	}
	return network_value, true
}

// Sanity checks for kill command and returns the relevant flags
func CheckKillCommand(kill_command string) (int, bool) {
	// Disregards the first word "kill"
	key_value := strings.Split(kill_command, " ")[1:]
	if len(key_value) != 2 || key_value[0] != "--id" {
		fmt.Println("Please enter the kill command following the convention of:\n" +
			"kiill --id id")
		return math.MinInt, false
	}
	// Makes sure id is within the range of [1, 65535]. ParseInt, with bitSize 16 checks the
	// upper bound. More can be found on tinyurl.com/yckj2twn.
	if id, err := strconv.ParseInt(key_value[1], 10, 64); err != nil {
		fmt.Println("Failed to parse id: ", err)
	} else {
		if id < 1 || id > 65535 {
			fmt.Println("Please enter a id in the range of [1, 65535].")
			return math.MinInt, false
		}
		return int(id), true
	}
	return math.MinInt, false
}

func CheckPlayExpertCommand(playexpert_command string, running_agents_num int64) map[string]float64 {
	return_map := make(map[string](float64))
	// Disregard the first word "playexpert"
	flag_list := strings.Split(playexpert_command, " ")[1:]

	// The flag list for playexpert command should contain the following elements. The key-value
	// pair can be in a different order. The length of the list thus needs to be 4.
	// [--num-agents number --liar-ratio ratio]
	if len(flag_list) != 4 {
		return nil
	}
	for i, element := range flag_list {
		switch element {
		case "--num-agents":
			if (i + 1) >= len(flag_list) {
				return nil
			}
			num_agents, valid := CheckNumAgentsFlag(flag_list[i+1], running_agents_num)
			if !valid {
				return nil
			}
			return_map["num_agents"] = num_agents

		case "--liar-ratio":
			if (i + 1) >= len(flag_list) {
				return nil
			}
			liar_ratio, valid := CheckLiarRatioFlag(flag_list[i+1])
			if !valid {
				return nil
			}
			return_map["liar_ratio"] = liar_ratio
		default:
			continue
		}
	}
	// Only if both two flags are found does this function return the map; otherwise,
	// that means there are some unidentified flags, thus causing an incomplete map.
	if len(return_map) == 2 {
		return return_map
	}
	return nil
}

// Sanity checks for start or extend command and returns the relevant flags
func CheckStartOrExtendCommand(start_command string) map[string]float64 {
	return_map := make(map[string]float64)
	// Disregards the first word "start" or "extend"
	flag_list := strings.Split(start_command, " ")[1:]

	// The flag list for start/extend command should contain the following elemnts. The key-value
	// pair can be in a different order. The length of the list thus needs to be 8.
	// [--value, v, --max-value, max, --num-agents, number, --liar-ratio, ratio]
	if len(flag_list) != 8 {
		return nil
	}
	for i, element := range flag_list {
		switch element {
		case "--value":
			if (i + 1) >= len(flag_list) {
				return nil
			}
			if value, err := strconv.ParseInt(flag_list[i+1], 10, 64); err == nil {
				return_map["value"] = float64(value)
			} else {
				fmt.Println("Error when parsing value: ", err)
				return nil
			}
		case "--max-value":
			if (i + 1) >= len(flag_list) {
				return nil
			}
			if max_value, err := strconv.ParseInt(flag_list[i+1], 10, 64); err == nil {
				if max_value < 1 {
					fmt.Println("max_value must be an integer >= 1")
					return nil
				}
				return_map["max_value"] = float64(max_value)
			} else {
				fmt.Println("Error when parsing max_value: ", err)
				return nil
			}
		case "--num-agents":
			if (i + 1) >= len(flag_list) {
				return nil
			}
			num_agents, valid := CheckNumAgentsFlag(flag_list[i+1], math.MaxInt64)
			if !valid {
				return nil
			}
			return_map["num_agents"] = num_agents

		case "--liar-ratio":
			if (i + 1) >= len(flag_list) {
				return nil
			}
			liar_ratio, valid := CheckLiarRatioFlag(flag_list[i+1])
			if !valid {
				return nil
			}
			return_map["liar_ratio"] = liar_ratio

		default:
			continue
		}
	}
	// Only if all four flags are found does this function return the map; otherwise,
	// that means there are some unidentified flags, thus causing an incomplete map.
	if len(return_map) == 4 {
		if return_map["max_value"] == 1 && return_map["value"] == return_map["max_value"] {
			fmt.Println("network_value and max_value cannot both be equal to 1 " +
				"because fake arbitrary value x needs to be 1 <= x <= max_value and" +
				"x != network_value")
			return nil
		}
		return return_map
	}
	return nil
}

func CheckLiarRatioFlag(ratio string) (float64, bool) {
	liar_ratio, err := strconv.ParseFloat(ratio, 64)
	if err == nil {
		if liar_ratio < 0 || liar_ratio > 1 {
			fmt.Println("liar_ratio must be >= 0 and <= 1")
			return math.MaxFloat64, false
		}
		return liar_ratio, true
	}
	fmt.Println("Error when parsing liar_ratio: ", err)
	return math.MaxFloat64, false
}

func CheckNumAgentsFlag(num_agents_str string, running_agents_num int64) (float64, bool) {
	// Parsing it to because that is the max of ports number
	// More can be found tinyurl.com/yckj2twn.
	num_agents, err := strconv.ParseInt(num_agents_str, 10, 64)
	if err == nil {
		if num_agents > 65535 {
			fmt.Println("There can be 65535 agents at most.")
			return math.MaxFloat64, false
		}
		if num_agents <= 0 || num_agents > running_agents_num {
			fmt.Println("num_agents must be an integer >= 1.",
				"For playexpert command, num_agents also needs to be  <= the number of runing agents.")
			return math.MaxFloat64, false
		}
		return float64(num_agents), true
	}
	fmt.Println("Error when parsing num_agents: ", err)
	return math.MaxFloat64, false
}
