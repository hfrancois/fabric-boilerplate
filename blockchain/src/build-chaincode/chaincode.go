package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"os"
	"build-chaincode/util"
	"build-chaincode/entities"
)

var logger = shim.NewLogger("fabric-boilerplate")
//======================================================================================================================
//	 Structure Definitions
//======================================================================================================================
//	SimpleChaincode - A blank struct for use with Shim (An IBM Blockchain included go file used for get/put state
//					  and other IBM Blockchain functions)
//==============================================================================================================================
type Chaincode struct {
}

//======================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Passes the
//  		 initial arguments passed are passed on to the called function.
//======================================================================================================================

func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface, functionName string, args []string) ([]byte, error) {
	logger.Infof("Invoke is running " + functionName)

	if functionName == "init" {
		return t.Init(stub, "init", args)
	} else if functionName == "resetIndexes" {
		return nil, util.ResetIndexes(stub, logger)
	} else if functionName == "addUser" {
		return nil, t.addUser(stub, args[0], args[1])
	} else if functionName == "addTestdata" {
		return nil, t.addTestdata(stub, args[0], args[1], args[2])
	} else if functionName == "createThing" {
		thingAsJSON := args[0]

		var thing entities.Thing
		if err := json.Unmarshal([]byte(thingAsJSON), &thing); err != nil {
			return nil, errors.New("Error while unmarshalling thing, reason: " + err.Error())
		}

		thingAsBytes, err := json.Marshal(thing);
		if err != nil {
			return nil, errors.New("Error marshalling thing, reason: " + err.Error())
		}

		util.StoreObjectInChain(stub, thing.ThingID, util.ThingsIndexName, thingAsBytes)

		return nil, nil
	} else if functionName == "addProject" {
		
		return nil, t.addProject(stub, args[0])

	} else if functionName == "addVoter" {
		
		return nil, t.addVoter(stub, args[0])

	} else if functionName == "vote" {
		
		return nil, t.vote(stub, args[0])

	} 

	return nil, errors.New("Received unknown invoke function name")
}

//======================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *Chaincode) Query(stub shim.ChaincodeStubInterface, functionName string, args []string) ([]byte, error) {
	logger.Infof("Query is running " + functionName)

	result, err := t.GetQueryResult(stub, functionName, args)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (t *Chaincode) GetQueryResult(stub shim.ChaincodeStubInterface, functionName string, args []string) (interface{}, error) {
	if functionName == "getUser" {
		user, err := util.GetUser(stub, args[0])
		if err != nil {
			return nil, err
		}

		return user, nil
	} else if functionName == "authenticateAsUser" {
		user, err := util.GetUser(stub, args[0])
		if err != nil {
			logger.Infof("User with id %v not found.", args[0])
		}

		return t.authenticateAsUser(stub, user, args[1]), nil
	} else if functionName == "getThingsByUserID" {
		thingsByUserID, err := util.GetThingsByUserID(stub, args[0])
		if err != nil {
			return nil, errors.New("could not retrieve things by user id: " + args[0] + ", reason: " + err.Error())
		}

		return thingsByUserID, nil
		
	} else if functionName == "getProjectsForVoter" {

		projects, err := util.GetProjectsForVoter(stub, args[0])
		
		if err != nil {
			return nil, errors.New("could not retrieve projects, reason: " + err.Error())
		}

		return projects, nil
		
	} else if functionName == "getVoter" {

		voter, err := util.GetVoter(stub, args[0])
		
		if err != nil {
			return nil, errors.New("could not retrieve voter, reason: " + err.Error())
		}

		return voter, nil
		
	} else if functionName == "getVoteForProjectByVoter" {

		voter, err := util.GetVoteForProjectByVoter(stub, args[0], args[1])
		
		if err != nil {
			return nil, errors.New("could not retrieve vote for project" + args[0] + ", voterId = " + args[1] + "reason: " + err.Error())
		}

		return voter, nil
		
	}

	return nil, errors.New("Received unknown query function name")
}

//======================================================================================================================
//  Main - main - Starts up the chaincode
//======================================================================================================================

func main() {
	// LogDebug, LogInfo, LogNotice, LogWarning, LogError, LogCritical (Default: LogDebug)
	logger.SetLevel(shim.LogInfo)

	logLevel, _ := shim.LogLevel(os.Getenv("SHIM_LOGGING_LEVEL"))
	shim.SetLoggingLevel(logLevel)

	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode: %s", err)
	}
}

//======================================================================================================================
//  Init Function - Called when the user deploys the chaincode
//======================================================================================================================

func (t *Chaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

//======================================================================================================================
//  Invoke Functions
//======================================================================================================================

func (t *Chaincode) addProject(stub shim.ChaincodeStubInterface, projectJSONObject string) error {
	
		var project entities.Project
		
		if err := json.Unmarshal([]byte(projectJSONObject), &project); err != nil {
			return errors.New("Error while unmarshalling project, reason: " + err.Error())
		}

		projectAsBytes, err := json.Marshal(project);
		
		if err != nil {
			return errors.New("Error marshalling project, reason: " + err.Error())
		}

		util.StoreObjectInChain(stub, project.ProjectID, util.ProjectsIndexName, projectAsBytes)
		
		//also store empty array of votes for this project
		emptyVotesAsBytes, errVotesMarshalling := json.Marshal([]entities.Vote{});
		
		if errVotesMarshalling != nil {
			return errors.New("Error marshalling []entities.Vote{}, reason: " + err.Error())
		}
		
		projectVoteKey := util.ProjectVotePrefix + project.ProjectID
		logger.Infof("Initializing emtpy votes for key = "+ projectVoteKey + ", emptyVotesAsBytes = " + string(emptyVotesAsBytes))
		stub.PutState(projectVoteKey, emptyVotesAsBytes)
		
		return nil
	
}

func (t *Chaincode) addVoter(stub shim.ChaincodeStubInterface, voterJSONObject string) error {
	
		var voter entities.Voter
		
		if err := json.Unmarshal([]byte(voterJSONObject), &voter); err != nil {
			return errors.New("Error while unmarshalling voter, reason: " + err.Error())
		}

		voterAsBytes, err := json.Marshal(voter);
		
		if err != nil {
			return errors.New("Error marshalling voter, reason: " + err.Error())
		}

		//util.StoreObjectInChain(stub, voter.VoterId, util.VotersIndexName, voterAsBytes)
		util.StoreObjectInChain(stub, util.VoterIndexPrefix + voter.VoterId, util.VotersIndexName, voterAsBytes)

		return nil
	
}

func (t *Chaincode) vote(stub shim.ChaincodeStubInterface, voteJSONObject string) error {
	
		logger.Infof("Voting START voteJSONObject=" + voteJSONObject)

		var vote entities.Vote
		
		if err := json.Unmarshal([]byte(voteJSONObject), &vote); err != nil {
			return errors.New("Error while unmarshalling vote, reason: " + err.Error())
		}

		logger.Infof("Marshalling vote")
		voteAsBytes, err := json.Marshal(vote);
		logger.Infof("Marshalled vote")
		
		if err != nil {
			return errors.New("Error marshalling vote, reason: " + err.Error())
		}
		
		//validate the vote
		logger.Infof("Validating the vote")
		projectAsBytes, err := stub.GetState(vote.ProjectID)
		if err != nil {
			return errors.New("Could not retrieve project for ID " + vote.ProjectID + " reason: " + err.Error())
		}

		var project entities.Project
		err = json.Unmarshal(projectAsBytes, &project)
		if err != nil {
			return errors.New("Error while unmarshalling projectAsBytes when voting, reason: " + err.Error())
		}

		
		logger.Infof("Will validate vote for project " + string(projectAsBytes))
		if !util.ValidateProjectForVoterId(stub, project, vote.VoterId) {
			return errors.New("Voter is not allowed to vote!")
		}
		
		//vote has been validated
		voteKey := util.VoteIndexPrefix + "_" + vote.VoterId + "_" + vote.ProjectID
		logger.Infof("Vote valid. Will save it in blockchain with key " + voteKey)
		
		//save vote on the blockchain
		util.StoreObjectInChain(stub, voteKey, util.VotesIndexName, voteAsBytes)
		logger.Infof("Saved vote in blockchain with key " + voteKey)
		
		//save projectvote
		//get existing votes for this project
		existingVotesForProject, err := util.GetVotesByProjectID(stub, vote.ProjectID)
		if err != nil {
			return errors.New("Error while getting existingVotesForProject, reason: " + err.Error())
		}
		logger.Infof("Found existingVotesForProject with length =%d " , len(existingVotesForProject))
		
		existingVotesForProject = append(existingVotesForProject, vote)
		logger.Infof("New existingVotesForProject with length =%d " , len(existingVotesForProject))
		
		existingVotesForProjectAsBytes, err := json.Marshal(existingVotesForProject);
		if err != nil {
			return errors.New("Error while marshalling existingVotesForProject, reason: " + err.Error())
		}
		stub.PutState(util.ProjectVotePrefix + vote.ProjectID, existingVotesForProjectAsBytes)
		logger.Infof("Added current vote to existing votes for this project. existingVotesForProjectAsBytes = " + string(existingVotesForProjectAsBytes))

		
		//update project costCovered
		project.CostCovered = project.CostCovered + vote.VotePercent
		projectAsBytes, err = json.Marshal(project);
		if err != nil {
			return errors.New("Could not marshal project after updating cost covered reason: " + err.Error())
		}
		stub.PutState(project.ProjectID, projectAsBytes)
		logger.Infof("Updated costCovered for project. projectAsBytes = " + string(projectAsBytes))

		return nil
	
}



func (t *Chaincode) addUser(stub shim.ChaincodeStubInterface, index string, userJSONObject string) error {
	id, err := util.WriteIDToBlockchainIndex(stub, util.UsersIndexName, index)
	if err != nil {
		return errors.New("Error creating new id for user " + index)
	}

	err = stub.PutState(string(id), []byte(userJSONObject))
	if err != nil {
		return errors.New("Error putting user data on ledger")
	}

	return nil
}

func (t *Chaincode) addTestdata(stub shim.ChaincodeStubInterface, testDataAsJson string, testVotersDataAsJson string, testProjectsDataAsJson string) error {
	var testData entities.TestData
	err := json.Unmarshal([]byte(testDataAsJson), &testData)
	if err != nil {
		return errors.New("Error while unmarshalling testdata")
	}

	for _, user := range testData.Users {
		userAsBytes, err := json.Marshal(user);
		if err != nil {
			return errors.New("Error marshalling testUser, reason: " + err.Error())
		}

		err = util.StoreObjectInChain(stub, user.UserID, util.UsersIndexName, userAsBytes)
		if err != nil {
			return errors.New("error in storing object, reason: " + err.Error())
		}
	}

	for _, thing := range testData.Things {
		thingAsBytes, err := json.Marshal(thing);
		if err != nil {
			return errors.New("Error marshalling testThing, reason: " + err.Error())
		}

		err = util.StoreObjectInChain(stub, thing.ThingID, util.ThingsIndexName, thingAsBytes)
		if err != nil {
			return errors.New("error in storing object, reason: " + err.Error())
		}
	}
	
	//voters data
	logger.Infof("testVotersDataAsJson = " + testVotersDataAsJson)

	var testVotersData []entities.Voter
	testVotersDataUnmarshallError := json.Unmarshal([]byte(testVotersDataAsJson), &testVotersData)
	if testVotersDataUnmarshallError != nil {
		return errors.New("Error while unmarshalling testVotersDataAsJson")
	}

	for _, testVoter := range testVotersData {
		testVoterAsBytes, err := json.Marshal(testVoter);
		if err != nil {
			return errors.New("Error marshalling testVoter, reason: " + err.Error())
		}

		util.StoreObjectInChain(stub, util.VoterIndexPrefix + testVoter.VoterId, util.VotersIndexName, testVoterAsBytes)

	}
	
	//projects data
	logger.Infof("testProjectsDataAsJson = " + testProjectsDataAsJson)

	var testProjectsData []entities.Project
	testProjectsDataUnmarshallError := json.Unmarshal([]byte(testProjectsDataAsJson), &testProjectsData)
	if testProjectsDataUnmarshallError != nil {
		return errors.New("Error while unmarshalling testProjectsDataAsJson")
	}

	for _, testProject := range testProjectsData {
		testProjectAsBytes, err := json.Marshal(testProject);
		if err != nil {
			return errors.New("Error marshalling testProject, reason: " + err.Error())
		}

		util.StoreObjectInChain(stub, testProject.ProjectID, util.ProjectsIndexName, testProjectAsBytes)

		//also store empty array of votes for this project
		emptyVotesAsBytes, errVotesMarshalling := json.Marshal([]entities.Vote{});
		
		if errVotesMarshalling != nil {
			return errors.New("Error marshalling []entities.Vote{}, reason: " + err.Error())
		}
		
		projectVoteKey := util.ProjectVotePrefix + testProject.ProjectID
		logger.Infof("Initializing emtpy votes for key = "+ projectVoteKey + ", emptyVotesAsBytes = " + string(emptyVotesAsBytes))
		stub.PutState(projectVoteKey, emptyVotesAsBytes)

	}
	

	return nil
}

//======================================================================================================================
//		Query Functions
//======================================================================================================================

func (t *Chaincode) authenticateAsUser(stub shim.ChaincodeStubInterface, user entities.User, passwordHash string) (entities.UserAuthenticationResult) {
	if user == (entities.User{}) {
		fmt.Println("User not found")
		return entities.UserAuthenticationResult{
			User: user,
			Authenticated: false,
		}
	}

	if user.Hash != passwordHash {
		fmt.Println("Hash does not match")
		return entities.UserAuthenticationResult{
			User: user,
			Authenticated: false,
		}
	}

	return entities.UserAuthenticationResult{
		User: user,
		Authenticated: true,
	}
}

