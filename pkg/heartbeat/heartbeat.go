package heartbeat

type HeartBeatRecord struct {
	SessionID  string `bson:"sessionid"`
	Scope      string `bson:"scope"`
	Region     string `bson:"region"`
	SubRegion  string `bson:"subregion"`
	HostName   string `bson:"hostname"`
	IP         string `bson:"ip"`
	MachineID  string `bson:"machineid"`
	LastHBUnix int64  `bson:"lasthbunix"`
	LastHB     string `bson:"lasthb"`
}

func main() {

}
