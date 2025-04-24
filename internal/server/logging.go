package server

const (
	StartingServerMsg         = "STARTING_SERVER"
	ServerClosedMsg           = "SERVER_CLOSED"
	ShutdownSignalMsg         = "SHUTDOWN_SIGNAL_RECEIVED"
	FailedShutdownMsg         = "FAILED_TO_SHUTDOWN_SERVER"
	ServerShutdownCompleteMsg = "SERVER_SHUTDOWN_COMPLETE"
	QueueShutdownSignalMsg    = "QUEUE_SHUTDOWN_SIGNAL_RECEIVED"
	QueueShutdownCompleteMsg  = "QUEUE_SHUTDOWN_COMPLETE"
	ProcessShutdownMsg        = "PROCESS_SHUTDOWN_COMPLETE"
	LogKeyError               = "error"
	LogKeyAddr                = "addr"
)
