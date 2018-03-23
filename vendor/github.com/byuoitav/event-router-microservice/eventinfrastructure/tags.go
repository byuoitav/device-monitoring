package eventinfrastructure

/*
In the context of our system there are five main players.
1. Local Event Generators
2. Room Event Proxies
3. External Event Translator
4. Local Event Consumers
5. Local Proxy (this microservice)


For the purposes of this microservice, all events flow through the local proxy.
Different event types have different routing rules to the different players in the system
*/

const Room = "room"

const UI = "ui"

const APISuccess = "api-success"

const APIError = "api-error"

const Translator = "translator"

const External = "external"

const Health = "health"

const Metrics = "metrics"

const UIFeature = "uifeature"

const RoomDivide = "roomdivide"

// for measuring status of event routing
const (
	TestStart         = "teststart"
	TestPleaseReply   = "testpleasereply"
	TestExternal      = "testexternal"
	TestExternalReply = "testexternalreply"
	TestReply         = "testreply"
	TestEnd           = "testend"
)
