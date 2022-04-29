package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GoooGu/liarslie/liars_network"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ModeType int64

const (
	STANDARD ModeType = iota
	EXPERT
)

func main() {
	mode_flag := flag.String("mode", "standard", "The mode in which the user wants to play.")
	flag.Parse()
	var curr_mode ModeType = STANDARD
	// Makes sure the mode can only be standard or expert
	if *mode_flag != "standard" && *mode_flag != "expert" {
		fmt.Println("Please select either standard or expert mode.")
	}
	if *mode_flag == "expert" {
		curr_mode = EXPERT
	}
	launched_agents_list := []*liars_network.Agent{}
	var honest_agents_num int
	rand.Seed(time.Now().UnixNano())
	scanner := bufio.NewScanner(os.Stdin)
command_reader_loop:
	for {
		if scanner.Scan() {
			command := scanner.Text()
			switch strings.Split(command, " ")[0] {
			case "start":
				if curr_mode != STANDARD {
					fmt.Println("Please only enter the available commands in expert mode: extend, playexpert & kill.")
					continue
				}
				if !LaunchAgents(&launched_agents_list, &honest_agents_num, command, curr_mode) {
					continue
				}
			case "play":
				if curr_mode != STANDARD {
					fmt.Println("Please only enter the available commands in expert mode: extend, playexpert & kill.")
					continue
				}
				if len(launched_agents_list) == 0 {
					fmt.Println("Please make sure you enter the start command first before you play.")
					continue
				}
				PlayCommand(honest_agents_num)

			case "stop":
				if curr_mode != STANDARD {
					fmt.Println("Please only enter the available commands in expert mode: extend, playexpert & kill.")
					continue
				}
				for _, agent := range launched_agents_list {
					agent.Stop()
				}
				fmt.Println("Deleting agents.config...")
				e := os.Remove("agents.config")
				if e != nil {
					log.Fatalln("Failed to delete agents.config: ", e)
				}
				break command_reader_loop

			case "extend":
				if curr_mode != EXPERT {
					fmt.Println("Please only enter the available commands in standard mode: start, play & stop.")
					continue
				}
				if !LaunchAgents(&launched_agents_list, &honest_agents_num, command, curr_mode) {
					continue
				}

			case "playexpert":
				if curr_mode != EXPERT {
					fmt.Println("Please only enter the available commands in standard mode: start, play & stop.")
					continue
				}
				if !PlayExpertCommand(honest_agents_num, launched_agents_list, command) {
					continue
				}

			case "kill":
				if curr_mode != EXPERT {
					fmt.Println("Please only enter the available commands in standard mode: start, play & stop.")
					continue
				}
				id, valid := liars_network.CheckKillCommand(command)
				// In case of an invalid kill command
				if !valid {
					continue
				}
				var is_agent_found bool = false
				// Linearly search for an agent whose port number matches the input id
				for i, agent := range launched_agents_list {
					if agent.IsMatchingPortNumber(id) {
						// Removes the agents from the list by swapping the agent about to be removed with
						// the agent at the end of the list and truncating the list to original_size - 1.
						original_size := len(launched_agents_list)
						launched_agents_list[i] = launched_agents_list[original_size-1]
						launched_agents_list = launched_agents_list[:original_size-1]
						agent.Stop()
						is_agent_found = true
						break
					}
				}
				if !is_agent_found {
					fmt.Println("Fails to find a matching agent whose id/port_number is ", id)
				}
			default:
				fmt.Println("Cannot recognize command:", scanner.Text())
			}
		}
	}
}

// Handles both extend and start command.
func LaunchAgents(launched_agents_list *[]*liars_network.Agent, honest_agents_num *int, command string, curr_mode ModeType) bool {
	if curr_mode == STANDARD && len(*launched_agents_list) != 0 {
		fmt.Println("The start command has already been run. You cannot rerun it.")
		return false
	}
	flags_map := liars_network.CheckStartOrExtendCommand(command)
	if flags_map == nil {
		if curr_mode == STANDARD {
			fmt.Println("Please enter the start command following the convention of:\n" +
				"start --value v --max-value max --num-agents number --liar-ratio ratio")
		} else {
			fmt.Println("Please enter the extend command following the convention of:\n" +
				"extend --value v --max-value max --num-agents number --liar-ratio ratio")
		}
		return false
	}
	network_value := int32(flags_map["value"])
	max_value := int32(flags_map["max_value"])
	new_agents_num := int(flags_map["num_agents"])
	liar_ratio := flags_map["liar_ratio"]

	// If called from start, then len(launched_agents_list) is always 0.
	// If called from extend, then len(launched_agents_list) could be 0 or non-zero.
	total_num_agents := len(*launched_agents_list) + new_agents_num
	liar_agents_num := int(liar_ratio * float64(total_num_agents))
	*honest_agents_num = total_num_agents - liar_agents_num

	// If in expert mode and agents.config does not exist, then creates a new agents.config file. If in standard mode,
	// always creates a new agents.config.
	var agents_config *os.File
	if _, err := os.Stat("agents.config"); curr_mode == EXPERT && errors.Is(err, os.ErrNotExist) || curr_mode == STANDARD {
		agents_config, err = os.Create("agents.config")
		if err != nil {
			log.Fatalf("Failed to create agents.config %s", err)
		}
	} else {
		agents_config, err = os.OpenFile("agents.config", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to read agent.config: %s", err)
		}
	}
	defer agents_config.Close()
	config_writer := csv.NewWriter(agents_config)
	for i := 0; i < total_num_agents; i++ {
		// Initializes agent_value to be the input network_value; updates it according to
		// how many liar_agents we have upped - if the number of liar_agents upped is below
		// the threshold, then modifies agent_value to an arbitrary number. Otherwise, keeps
		// it unchanged.
		var agent_value int32 = network_value
		if i < liar_agents_num {
			var arbitrary_value int32 = 1 + rand.Int31n(max_value)
			// Makes sure there is no collision between network value and arbitrary value.
			for arbitrary_value == int32(network_value) {
				arbitrary_value = 1 + rand.Int31n(max_value)
			}
			agent_value = arbitrary_value
		}
		// Creates a new agent
		if i < new_agents_num {
			var wait_group sync.WaitGroup
			wait_group.Add(1)
			port_number := make(chan int)
			new_agent := new(liars_network.Agent)
			*launched_agents_list = append(*launched_agents_list, new_agent)
			go new_agent.Init(port_number, agent_value, &wait_group)
			config_writer.Write([]string{strconv.FormatInt(int64(<-port_number), 10)})
			config_writer.Flush()
			wait_group.Wait()
		} else {
			// This condition should only be entered in EXPERT Mode. For the already launched agents,
			// updates their values to reflect the newly added agents and the input from the extend
			// command.
			fmt.Println("Existing agent ", i-new_agents_num, " updating its value...")
			(*launched_agents_list)[i-new_agents_num].UpdateValue(agent_value)
		}
	}
	fmt.Println("Ready")
	return true
}

// Handles play command in standard mode
func PlayCommand(honest_agents_num int) {
	// Tries to find the agents.config file and reads from it
	agents_config, err := os.Open("agents.config")
	if err != nil {
		log.Fatalf("Failed to open agents.config: %s", err)
	}
	agents_port_nums_list, err := csv.NewReader(agents_config).ReadAll()
	if err != nil {
		log.Fatalf("Failed to read agent.config: %s", err)
	}
	responses := []int32{}
	// Retrieves grpc responses from each client and then collects all the responses
	for _, row := range agents_port_nums_list {
		conn, err := grpc.Dial(":"+row[0], grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("could not connect: %s", err)
		}
		defer conn.Close()
		client := liars_network.NewLieServiceClient(conn)
		response, err := client.LieQuery(context.Background(), new(liars_network.LieRequest))
		if err != nil {
			log.Fatalf("Error when calling LieQuery: %s", err)
		}
		responses = append(responses, response.AgentValue)
	}
	// The network value is found by finding the unique element from the slice which matches the same frequncy,
	// which is the number of honest agents in the network. If there are more than one value whose frequency
	// matches the number of honest agents, then a correct network value cannot be decided.
	if network_value, exists := liars_network.FindNetworkValue(responses, honest_agents_num); exists {
		fmt.Println("The network value is ", network_value)
	} else {
		fmt.Println("The network value cannot be decided because the liar agents successfully fooled the client.")
	}
}

func PlayExpertCommand(honest_agents_num int, launched_agents_list []*liars_network.Agent, command string) bool {
	if len(launched_agents_list) == 0 {
		fmt.Println("Please make sure you enter the extend command first before you playexpert.")
		return false
	}
	flag_map := liars_network.CheckPlayExpertCommand(command, int64(len(launched_agents_list)))
	if flag_map == nil {
		fmt.Println("Please enter the playexpert command following the convention of:\n" + "playexpert --num-agents number --liar-ratio ratio")
		return false
	}
	liar_ratio := flag_map["liar_ratio"]
	// The frequency of the network value based on the assumption given from the user.
	assumed_frequency := len(launched_agents_list) - int(liar_ratio*float64(len(launched_agents_list)))
	if assumed_frequency != honest_agents_num {
		fmt.Println("Warning: the input of liar_ratio in playexpert differs from that of the most recent extend.")
	}
	proxy_agent := launched_agents_list[0]

	// The port numbers of all the agents which the proxy agent communicates with.
	var other_agent_ids []string
	for _, other_agent := range launched_agents_list[1:] {
		other_agent_ids = append(other_agent_ids, other_agent.RetrievePortNum())
	}

	// Establishes the connection with the proxy agent.
	conn, err := grpc.Dial(":"+proxy_agent.RetrievePortNum(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %s", err)
	}
	defer conn.Close()

	// Queries the proxy agent
	client := liars_network.NewLieServiceClient(conn)
	response, err := client.LieQuery(context.Background(), &liars_network.LieRequest{ExpertMode: true, OtherAgentIds: other_agent_ids})
	if err != nil {
		log.Fatalf("Error when calling LieQuery: %s", err)
	}

	// Append the value from the proxy agent to the collected values from the rest of the network
	all_values_from_network := response.GetCollectedAgentValues()
	all_values_from_network = append(all_values_from_network, response.AgentValue)

	// The network value is found by finding the unique element from the slice which matches the same frequncy,
	// which is the number of honest agents in the network. If there are more than one value whose frequency
	// matches the number of honest agents, then a correct network value cannot be decided.
	if network_value, exists := liars_network.FindNetworkValue(all_values_from_network, assumed_frequency); exists {
		fmt.Println("The network value is ", network_value)
	} else {
		fmt.Println("The network value cannot be decided because the liar agents successfully fooled the client.")
	}
	return true
}
