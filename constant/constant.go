package constant

const (
	//TODO: Use codes
	// UUIDMissing : UUID not passed in request
	UUIDMissing = "101"

	// InvalidUUID : UUID length is not 36 characters long which is the expected length
	InvalidUUID = "102"

	// InvalidJSON : Invalid JSON sent in request body
	InvalidJSON = "103"

	// KeyCountExceeded : more keys than allowed in request body
	KeyCountExceeded = "104"

	// MaxSizeExceeded : POST /cache element exceeded max size
	MaxSizeExceeded = "105"

	// TimedOut : Timeout writing value to the backend
	TimedOut = "106"

	// UnexpErr : POST /cache had an unexpected error
	UnexpErr = "107"
)

const (
	DefaultNodeName = "Default_Node" //DefaultNodeName is the default node name for K8s environment
	DefaultPodName  = "Default_Pod"  //DefaultPodName is the default pod name for K8s environment
	DefaultDCName   = "Default_DC"   //DefaultPodName is the default pod name for K8s environment
	EnvVarNodeName  = "MY_NODE_NAME" //EnvVarNodeName is the environment variable for node name in K8s environment
	EnvVarPodName   = "MY_POD_NAME"  //EnvVarPodName is the environment variable for pod name in K8s environment
	EnvVarDCName    = "CLUSTER_NAME" //EnvVarDCName is the environment variable for cluster name in k8s environment
)
