package drivers

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
)

/*
For more information of the commands for the RGA, please visit https://mmrc.caltech.edu/Kratos%20XPS/MKS%20RGA/ASCII%20Protocol%20User%20Manual%20SP1040016.100.pdf
*/
var (
	commandSuffix = "\n\r"
	delim = []byte("\r\n")
	commandEnd = []byte("\r\n\r\r")
	fieldRegex := `\S+`
	/*Returns a table of sensors that can be controlled*/
	sensors = "Sensors"
	/*
	ARGS:
	    - SerialNumber: example, LM70-00197021
	*/
	selectCmd = "Select"
	sensorState = "SensorState"
	info = "Info"
	eGains = "EGains"
	inletInfo = "InletInfo"
	rfInfo = "RFInfo"
	multiplierInfo = "MultiplierInfo"
	/*
	ARGS:
	    - SourceIndexZero: example, 0
	*/
	sourceInfo = "SourceInfo"
	/*
	ARGS:
	    - SourceIndexZero: example, 0
	*/
	detectorInfo = "DetectorInfo"
	/*
	RESPONSE:
	    - SummaryState - Enum: OFF, WARM-UP, ON, COOL-DOWN, BAD-EMISSION
	*/
	filamentInfo = "FilamentInfo"
	/*
	RESPONSE:
	    - Pressure - In Pascals, 0 if filament is off
	*/
	totalPressureInfo = "TotalPressureInfo"
	analogInputInfo = "AnalogInputInfo"
	analogOutputInfo = "AnalogOutputInfo"
	digitalInfo	= "DigitalInfo"
	rolloverInfo = "RolloverInfo"
	rVCInfo = "RVCInfo"
	cirrusInfo = "CirrusInfo"
	/*
	ARGS:
		- SourceIndex The 0 based index of the source parameters
		- DetectorIndex The 0 based index of the detector (0=Faraday, 1,2,3=Multiplier settings)
	*/
	pECal_Info = "PECal_Info"
	/*
	ARGS:
		- AppName String specifying the application name of the controlling application
		- Version String specifying the version of the controlling application
	EXAMPLE:
		Control "Process Eye Pro" "5.1"
	*/
	control = "Control"
	release = "Release" //defer Release()
	/*
	ARGS:
		- State Can be ‘On’ or ‘Off’
	EXAMPLE:
		FilamentControl On
	*/
	filamentControl = "FilamentControl"
	/*
	ARGS:
		- Number The filament number to select: 1 or 2
	EXAMPLE:
		FilamentSelect 2
	*/
	filamentSelect = "FilamentSelect"
	/*
	ARGS:
		- Time Number of seconds to keep the filaments on for.
	EXAMPLE:
		FilamentOnTime 200
	*/
	filamentOnTime = "FilamentOnTime"
	/*
	ARGS:
		- Name 			The name that the measurement should be called
		- StartMass 	The starting mass that should be scanned
		- EndMass 		The ending mass that should be scanned
		- PointsPerPeak Number of points to be measured across each mass
		- Accuracy 		Accuracy code to be used
		- EGainIndex 	Electronic Gain index
		- SourceIndex 	Source parameters index
		- DetectorIndex Detector parameters index
	EXAMPLE:
		AddAnalog Analog1 1 50 32 5 0 0 0
	*/
	addAnalog = "AddAnalog"
	/*
	ARGS:
		- Name 			The name that the measurement should be called
		- StartMass 	The starting mass that should be scanned
		- EndMass 		The ending mass that should be scanned
		- FilterMode 	How masses should be scanned and converted into a single reading (ENUM: PeakCenter, PeakMax, PeakAverage)
		- Accuracy 		Accuracy code to be used
		- EGainIndex 	Electronic Gain index
		- SourceIndex 	Source parameters index
		- DetectorIndex Detector parameters index
	EXAMPLE:
		AddBarchart Bar1 1 50 PeakCenter 5 0 0 0
	*/
	addBarchart = "AddBarchart"
	/*
	ARGS:
		- Name 			The name that the measurement should be called
		- FilterMode 	How masses should be scanned and converted into a single reading (ENUM: PeakCenter, PeakMax, PeakAverage)
		- Accuracy 		Accuracy code to be used
		- EGainIndex 	Electronic Gain index
		- SourceIndex 	Source parameters index
		- DetectorIndex Detector parameters index
	EXAMPLE:
		AddPeakJump PeakJump1 PeakCenter 5 0 0 0
	*/
	addPeakJump = "AddPeakJump"
	/*
	ARGS:
		- Name 			The name that the measurement should be called
		- Mass 			The mass that should be measured
		- Accuracy 		Accuracy code to be used
		- EGainIndex 	Electronic Gain index
		- SourceIndex 	Source parameters index
		- DetectorIndex Detector parameters index
	EXAMPLE:
		AddSinglePeak SinglePeak1 4.2 5 0 0 0
	*/
	addSinglePeak = "AddSinglePeak"
	/*
	ARGS:
		- Accuracy 0 – 8 Accuracy code
	EXAMPLE:
		MeasurementAccuracy 4
	*/
	measurementAccuracy = "MeasurementAccuracy"
	/*
	ARGS:
		- Mass Integer mass value
	EXAMPLE:
		MeasurementAddMass 10
	*/
	measurementAddMass = "MeasurementAddMass"
	/*
	ARGS:
		- MassIndex Index of the mass that should be changed
		- NewMass 	New mass value that should be scanned instead
	EXAMPLE:
		MeasurementChangeMass 0 6
	*/
	measurementChangeMass = "MeasurementChangeMass"
	/*
	ARGS:
		- DetectorIndex 0 based index of the detector to use for the measurement
	EXAMPLE:
		MeasurementDetectorIndex 0
	*/
	measurementDetectorIndex ="MeasurementDetectorIndex"
	/*
	ARGS:
		- EGainIndex 0 based index of the electronic gain to use for the measurement
	EXAMPLE:
		MeasurementWGainIndex 1
	*/
	measurementEGainIndex = "MeasurementWGainIndex"
	/*
	ARGS:
		- FilterMode The mode to be used to filter readings down to 1 per AMU (ENUM: PeakCenter, PeakMax, PeakAverage)
	EXAMPLE:
		MeasurementFilterMode PeakCenter
	*/
	measurementFilterMode = "MeasurementFilterMode"
	/*
	ARGS:
		- Mass The mass value to use for the selected single peak measurement. Can be fractional
	EXAMPLE:
		MeasurementMass 15.5
	*/
	measurementMass = "MeasurementMass"
	/*
	ARGS:
		- PointsPerPeak The number of points per peak to be measured for Analog measurement
	EXAMPLE:
		MeasurementPointsPerPeak 16
	*/
	measurementPointsPerPeak = "MeasurementPointsPerPeak"
	/*
	ARGS:
		- MassIndex 0 based index of the mass peak to remove from a Peak Jump measurement
	EXAMPLE:
		MeasurementRemoveMass 1
	*/
	measurementRemoveMass = "MeasurementRemoveMass"
	/*
	ARGS:
		- SourceIndex 0 based index of the source parameters to use for the measurement
	EXAMPLE:
		MeasurementSourceIndex 0
	*/
	measurementSourceIndex = "MeasurementSourceIndex"
	/*
	ARGS:
		- UseCorrection True/False whether to use rollover correction for the selected measurement
	EXAMPLE:
		MeasurementRolloverCorrection True
	*/
	measurementRolloverCorrection = "MeasurementRolloverCorrection"
	/*
	ARGS:
		- BeamOff Boolean indicating if the beam should be off during zero readings.
	EXAMPLE:
		MeasurementZeroBeamOff True
	*/
	measurementZeroBeamOff = "MeasurementZeroBeamOff"
	/*
	ARGS:
		- ZeroBufferDepth The depth of the zero reading buffer.
	EXAMPLE:
		MeasurementZeroBufferDepth 8
	*/
	measurementZeroBufferDepth = "MeasurementZeroBufferDepth"
	/*
	ARGS:
		- ZeroBufferMode The mode of operation for the zero averaging logic (ENUM: SingleScanAverage, MultiScanAverage, MultiScanAverageQuickStart, SingleShot)
	EXAMPLE:
		MeasurementZeroBufferMode MultiScanAverage
	*/
	measurementZeroBufferMode = "MeasurementZeroBufferMode"
	measurementZeroReTrigger = "MeasurementZeroReTrigger"
	/*
	ARGS:
		- ZeroMass The mass value that should be used to take the zero readings for the measurement
	EXAMPLE:
		MeasurementZeroMass 5.5
	*/
	measurementZeroMass = "MeasurementZeroMass"
	/*
	ARGS:
		- Protect Boolean indicating if the multiplier should be locked by software.
	EXAMPLE:
		MultiplierProtect True
	*/
	multiplierProtect = "MultiplierProtect"
	runDiagnostics = "RunDiagnostics"
	/*
	ARGS:
		- Pressure Value to be used for total pressure [Pa]
	EXAMPLE:
		TotalPressure 1.0E-4
	*/
	setTotalPressure = "TotalPressure"
	/*
	ARGS:
		- Factor Float value to apply to total pressure reading from an external gauge
	EXAMPLE:
		TotalPressureCalFactor 1.0
	*/
	totalPressureCalFactor = "TotalPressureCalFactor"
	/*
	ARGS:
		- DateTime Date in form yyyy-mm-dd_HH:MM:SS
	EXAMPLE:
		TotalPressureCalDate 2005-10-06_16:44:00
	*/
	totalPressureCalDate = "TotalPressureCalDate"
	/*
	ARGS:
		- InletOption 		How to apply inlet calibration factor (ENUM: Off, Default, Current)
		- DetectorOption 	How to apply detector calibration factor (ENUM: Off, Default, Current)
	EXAMPLE:
		CalibrationOptions Off Off
	*/
	calibrationOptions = "CalibrationOptions"
	/*
	ARGS:
		- SourceIndex 	The 0 based index of the source settings being used
		- DetectorIndex The 0 based index of the detector settings being used
		- Filament 		The filament number 1 or 2. Or 0 if both filaments factors to be set
		- Factor 		The new calibration factor
	EXAMPLE:
		DetectorFactor 0 0 1 1.5e-6
	*/
	detectorFactor = "DetectorFactor"
	/*
	ARGS:
		- SourceIndex 	The 0 based index of the source settings being used
		- DetectorIndex The 0 based index of the detector settings being used
		- Filament 		The filament number 1 or 2. Or 0 if both filaments factors to be set
		- Date 			The time and date formatted as yyyy-mm-dd_HH:MM:SS
	EXAMPLE:
		DetectorCalDate 0 0 0 2005-06-01_11:49:00
	*/
	detectorCalDate = "DetectorCalDate"
	/*
	ARGS:
		- SourceIndex 	The 0 based index of the source settings being used
		- DetectorIndex The 0 based index of the detector settings being used
		- Filament 		The filament number 1 or 2. Or 0 if both filaments factors to be set
		- Voltage 		The new multiplier voltage to use
	EXAMPLE:
		DetectorVoltage 0 1 1 500
	*/
	detectorVoltage = "DetectorVoltage"
	/*
	ARGS:
		- InletIndex 	0 based index of the inlet to set the factor for.
		- Factor 		The new inlet factor
	EXAMPLE:
		InletFactor 0 1.5
	*/
	inletFactor = "InletFactor"
	/*
	ARGS:
		- MeasurementName The measurement to add to the scan
	EXAMPLE:
		ScanAdd Analog1
	*/
	scanAdd = "ScanAdd"
	/*
	ARGS:
		- NumScans Starts a scan running and will re-trigger the scan automatically the number of times specified by NumScans
	EXAMPLE:
		ScanStart 1
	*/
	scanStart = "ScanStart"
	scanStop = "ScanStop"
	/*
	ARGS:
		- NumScans Number of scans to re-trigger the scan for. (optional)
	EXAMPLE:
		ScanRestart
	*/
	scanResume = "ScanResume"
	/*
	ARGS:
		- NumScans Number of scans to re-trigger the scan for. (optional)
	EXAMPLE:
		ScanRestart
	*/
	scanRestart = "ScanRestart"
	/*
	ARGS:
		- MeasurementName The measurement that should be selected for other MeasurementXXXX commands
	EXAMPLE:
		MeasurementSelect Analog1
	*/
	measurementSelect = "MeasurementSelect"
	/*
	ARGS:
		- Mass The new start mass for the Analog or Barchart measurement
	EXAMPLE:
		MeasurementStartMass 50
	*/
	measurementStartMass = "MeasurementStartMass"
	/*
	ARGS:
		- Mass The new start mass for the Analog or Barchart measurement
	EXAMPLE:
		MeasurementEndMass 45
	*/
	measurementEndMass = "MeasurementEndMass"
	measurementRemoveAll = "MeasurementRemoveAll"
	/*
	ARGS:
		- MeasurementName Name of the measurement to remove
	EXAMPLE:
		MeasurementRemove Barchart1
	*/
	measurementRemove = "MeasurementRemove"
	/*
	ARGS:
		- UseTab Boolean indicating whether to use tab characters in the output or spaces.
	EXAMPLE:
		FormatWithTab True
	*/
	formatWithTab = "FormatWithTab"
	/*
	ARGS:
		- SourceIndex 	0 based index of the source parameters entry to modify
		- IonEnergy 	New ion energy value [eV]
	EXAMPLE:
		SourceIonEnergy 0 5.5
	*/
	sourceIonEnergy = "SourceIonEnergy"
	/*
	ARGS:
		- SourceIndex 	0 based index of the source parameters entry to modify
		- Emission 		New emission value. [mA]
	EXAMPLE:
		SourceEmission 0 1.0
	*/
	sourceEmission = "SourceEmission"
	/*
	ARGS:
		- SourceIndex 	0 based index of the source parameters entry to modify
		- Extract 		New extract value. [V]
	EXAMPLE:
		SourceExtract 0 -112
	*/
	sourceExtract = "SourceExtract"
	/*
	ARGS:
		- SourceIndex 		0 based index of the source parameters entry to modify
		- ElectronEnergy 	New electron energy value [eV]
	EXAMPLE:
		SourceElectronEnergy 0 70
	*/
	sourceElectronEnergy = "SourceElectronEnergy"
	/*
	ARGS:
		- SourceIndex 		0 based index of the source parameters entry to modify
		- LowMassResolution New low mass resolution value (0 - 65535)
	EXAMPLE:
		SourceLowMassResolution 0 32767
	*/
	sourceLowMassResolution = "SourceLowMassResolution"
	/*
	ARGS:
		- SourceIndex 		0 based index of the source parameters entry to modify
		- LowMassAlignment 	New low mass alignment value (0 - 65535)
	EXAMPLE:
		SourceLowMassAlignment 0 32767
	*/
	sourceLowMassAlignment = "SourceLowMassAlignment"
	/*
	ARGS:
		- SourceIndex 		0 based index of the source parameters entry to modify
		- HighMassAlignment New high mass alignment value (0 - 65535)
	EXAMPLE:
		SourceHighMassAlignment 0 32767
	*/
	sourceHighMassAlignment = "SourceHighMassAlignment"
	/*
	ARGS:
		- SourceIndex 			0 based index of the source parameters entry to modify
		- HighMassResolution 	New high mass resolution value (0 - 65535)
	EXAMPLE:
		SourceHighMassResolution 0 32767
	*/
	sourceHighMassResolution = "SourceHighMassResolution"
	/*
	ARGS:
		- Index 			The index of the analog input (Optional)
		- NumberToAverage 	The number of readings that should be averaged before returning result (Optional)
	EXAMPLE:
		AnalogInputAverageCount
	*/
	analogInputAverageCount = "AnalogInputAverageCount"
	/*
	ARGS:
		- Index 	The index of the analog input (Optional)
		- Enable 	True/False to enable or disable the analog input (Optional)
	EXAMPLE:
		AnalogInputEnable
	*/
	analogInputEnable = "AnalogInputEnable"
	/*
	ARGS:
		- Index 	The index of the analog input (Optional)
		- Interval 	Time in microseconds between successive analog input readings [µs] (Optional)
	EXAMPLE:
		AnalogInputInterval
	*/
	analogInputInterval = "AnalogInputInterval"
	/*
	ARGS:
		- Index 	The index of the analog output (Optional)
		- Value 	The value to set the analog output to (Optional)
	EXAMPLE:
		AnalogOutput
	*/
	analogOutput = "AnalogOutput"
	/*
	ARGS:
		- Frequency The frequency in Hz to drive the sensors audio output [Hz]
	EXAMPLE:
		AudioFrequency 1000
	*/
	audioFrequency = "AudioFrequency"
	/*
	ARGS:
		- Mode The mode to run the audio in. (ENUM: Off, Automatic, Manual)
	EXAMPLE:
		AudioMode Manual
	*/
	audioMode = "AudioMode"
	/*
	ARGS:
		- HeatOn True/False to turn heater on/off
	EXAMPLE:
		CirrusCapillaryHeater False
	*/
	cirrusCapillaryHeater = "CirrusCapillaryHeater"
	/*
	ARGS:
		- Mode State to put heater into: Off, Warm or Bake
	EXAMPLE:
		CirrusHeater Warm
	*/
	cirrusHeater = "CirrusHeater"
	/*
	ARGS:
		- PumpOn True/False to turn pump On/Off
	EXAMPLE:
		CirrusPump False
	*/
	cirrusPump = "CirrusPump"
	/*
	ARGS:
		- ValvePos 0 based valve position
	EXAMPLE:
		CirrusValvePosition 1
	*/
	cirrusValvePosition = "CirrusValvePosition"
	/*
	ARGS:
		- Time Time in seconds for port B bits 6 and 7 to remain set [s]
	EXAMPLE:
		DigitalMaxPB67OnTime 600
	*/
	digitalMaxPB67OnTime = "DigitalMaxPB67OnTime"
	/*
	ARGS:
		- Port 	The port name,A, B, C, etc.
		- Value The value to set outputs to. 8 bit number (0 – 255)
	EXAMPLE:
		DigitalOutput A 192
	*/
	digitalOutput = "DigitalOutput"
	/*
	ARGS:
		- Date 		in yyyy-mm-dd_HH:MM:SS format
		- Message 	Text message to be displayed when calibration is run
	EXAMPLE:
		PECal_DateMsg 2021-10-21_10:21:00 "Y'arr maties"
	*/
	pECal_DateMsg = "PECal_DateMsg"
	pECal_Flush = "PECal_Flush"
	/*
	ARGS:
		- Inlet1
		- Inlet2
		- Inlet3
	EXAMPLE:
		PECal_Inlet 1.0 1.0 1.0
	*/
	pECal_Inlet = "PECal_Inlet"
	/*
	ARGS:
		- Mass
		- Method
		- Contribution
	EXAMPLE:
		PECal_MassMethodContribution 28 0 80.5
	*/
	pECal_MassMethodContribution = "PECal_MassMethodContribution"
	pECal_Pressures = "PECal_Pressures"
	/*
	ARGS:
		- SourceIndex
		- DetectorIndex
	EXAMPLE:
		PECal_Select 0 0
	*/
	pECal_Select = "PECal_Select"
	/*
	ARGS:
		- Mass 		The mass to set a specific peak scale factor for
		- Factor 	The peak scale factor for the mass
	EXAMPLE:
		RolloverScaleFactor 28 5.2
	*/
	rolloverScaleFactor = "RolloverScaleFactor"
	/*
	ARGS:
		- M1
		- M2
		- B1
		- B2
		- BP1	
	EXAMPLE:
		RolloverVariables -470 -250 -0.15 -0.91 0.0012
	*/
	rolloverVariables = "RolloverVariables"
	/*
	ARGS:
		- State True/False value whether the alarm output should be set on or off.	
	EXAMPLE:
		RVCAlarm True
	*/
	rVCAlarm = "RVCAlarm"
	rVCCloseAllValves = "RVCCloseAllValves"
	/*
	ARGS:
		- HeaterOn True/False value whether to switch the heater on/off.	
	EXAMPLE:
		RVCHeater True
	*/
	rVCHeater = "RVCHeater"
	/*
	ARGS:
		- PumpOn True/False value whether to switch the pump on/off.	
	EXAMPLE:
		RVCPump True
	*/
	rVCPump = "RVCPump"
	/*
	ARGS:
		- Valve Index of the valve to open/close. (0 - 2)
		- Open 	True/False to open or close the valve	
	EXAMPLE:
		RVCValveControl 0 True
	*/
	rVCValveControl = "RVCValveControl"
	/*
	ARGS:
		- Mode Manual or Automatic	
	EXAMPLE:
		RVCValveMode Manual
	*/
	rVCValveMode = "RVCValveMode"
	saveChanges = "SaveChanges"
	/*
	ARGS:
		- StartPower 		Percentage power to start at. Typically 10% [%]
		- EndPower 			Percentage power to ramp to. Typically 85% [%]
		- RampPeriod 		Time in seconds to ramp between StartPower and EndPower. Typically 90s [s]
		- MaxPowerPeriod 	Time to hold at EndPower. Typically 240s [s]
		- ResettlePeriod 	Time to return to default settings. Typically 30s [s]
	EXAMPLE:
		StartDegas 10 85 90 240 30
	*/
	startDegas = "StartDegas"
	stopDegas = "StopDegas"
	BUFFER = 4096
	ACKMsg = "MKSRGA"
)

type RGAConnection net.Conn

type RGAErr string
type RGAType int32

const (
	Rga_ERROR RGAErr = "ERROR"
	Rga_OK RGAErr = "OK"
	Rga_INT RGAType = 0
	Rga_FLOAT RGAType = 1
	Rga_BOOL RGAType = 2
	Rga_STR RGAType = 3
)

type RGARespErr struct {
	CommandName string
	Err			RGAErr
}

type RGAValue struct {
	Type	RGAType
	Value	interface{}
}

type RGAResponse struct {
	ErrMsg	RGARespErr
	Fields	map[string]RGAValue
}

// parseHorizontalResp will parse an horizontal response from the RGA into a RGAResponse object
func parseHorizontalResp(resp []byte) (*RGAResponse, error) {
	re, err := regexp.Compile(fieldRegex)
	if err != nil {
		return nil, err
	}
	split := bytes.Split(resp, delim)
	//We know the first row is always the name of the command it's error status
	errorStatus := re.FindAllString(string(split[0]), 2)
	var errMsg error
	if RGAErr{errorStatus[1]} != Rga_ERROR && RGAErr{errorStatus[1]} != Rga_OK {
		return nil, fmt.Errorf("Unkown RGA error code: %s", errorStatus[1])
	} else {
		errMsg = fmt.Errorf(errorStatus[1])
	}
	// We know that the second row will be headers
	headers := re.FindAllString(string(split[1]), -1)
	// We know that split has a length of four since it's an horizontal response. We can ignore the last line.
	for j := 2; j < len(split)-1; j++ {
		values := re.FindAllString(string(split[j]), len(headers))
		fields := make(map[string]RGAValue)
		for i, header := range headers {
			if j > 2 {
				header = header+strconv.Itoa(i-2)
			}
			if _, ok := fields[header]; !ok {
				// if int64
				if v, err := strconv.ParseInt(values[i], 10, 64); err == nil {
					fields[header] = RGAValue{Type: Rga_INT, Value: v}
				// if float64
				} else if v, err := strconv.ParseFloat(values[i], 64); err == nil {
					fields[header] = RGAValue{Type: Rga_FLOAT, Value: v}
				// if bool
				} else if v, err := strconv.ParseBool(values[i]); err == nil {
					fields[header] = RGAValue{Type: Rga_BOOL, Value: v}
				// if string
				} else {
					fields[header] = RGAValue{Type: Rga_STR, Value: values[i]}
				}
			}
		}
	}
	
	return &RGAResponse{
		ErrMsg: RGAErr{errorStatus[1]},
		Fields: fields,
	}, errMsg
}

// parseVerticalResponse will parse a vertical response from the RGA into a RGAResponse object
func parseVerticalResponse(resp []byte, oneValuePerLine bool) (*RGAResponse, error) {
	re, err := regexp.Compile(fieldRegex)
	if err != nil {
		return nil, err
	}
	split := bytes.Split(resp, delim)
	//We know the first row is always the name of the command it's error status
	errorStatus := re.FindAllString(string(split[0]), 2)
	var errMsg error
	if RGAErr{errorStatus[1]} != Rga_ERROR && RGAErr{errorStatus[1]} != Rga_OK {
		return nil, fmt.Errorf("Unkown RGA error code: %s", errorStatus[1])
	} else {
		errMsg = fmt.Errorf(errorStatus[1])
	}
	// We know that the second to the before last will be our header value combos
	fields := make(map[string]RGAValue)
	for i := 1; i < len(split)-1; i++ {
		var (
			field []string
			name string
			value string
		)
		if oneValuePerLine {
			field = re.FindAllString(string(split[i]), 2)
			name = fmt.Sprintf("Value%d", i)
			value = field[0]
		} else {
			field = re.FindAllString(string(split[i]), 2)
			name = field[0]
			value = field[1]
		}
		if _, ok := fields[field[0]]; !ok {
			// if int64
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				fields[name] = RGAValue{Type: Rga_INT, Value: v}
			// if float64
			} else if v, err := strconv.ParseFloat(value, 64); err == nil {
				fields[name] = RGAValue{Type: Rga_FLOAT, Value: v}
			// if bool
			} else if v, err := strconv.ParseBool(value); err == nil {
				fields[name] = RGAValue{Type: Rga_BOOL, Value: v}
			// if string
			} else {
				fields[name] = RGAValue{Type: Rga_STR, Value: value}
			}
		}
	}
	return &RGAResponse{
		ErrMsg: RGAErr{errorStatus[1]},
		Fields: fields,
	}, errMsg
}

// InitMsg returns the standard response when the RGA receives it's first message from the client
func (c RGAConnection) InitMsg() error {
	fmt.Fprintf(c, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	split := bytes.Split(resp, delim)
	splitAgain := bytes.Split(split, delim)
	re, err := regexp.Compile(fieldRegex)
	if err != nil {
		return err
	}
	firstLine := re.FindAllString(string(splitAgain[0]), 2)
	if firstLine != ACKMsg {
		return fmt.Errorf("RGA did not respond with expected ACK msg: %s", string(split[0]))
	}
	return nil
}

// Sensors returns a table of sensors that can be controlled
func (c RGAConnection) Sensors() (*RGAResponse, error) {
	fmt.Fprintf(c, sensors+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseHorizontalResp(resp[0])
}

// Select selects the device via the inputted SerialNumber
func (c RGAConnection) Select(SerialNumber string) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %s%s", selectCmd, SerialNumber, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// SensorState retrieves the state of the selected sensor. State can only one of Ready, InUse, Config, N/A
func (c RGAConnection) SensorState() (*RGAResponse, error) {
	fmt.Fprintf(c, sensorState+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// Info returns the sensor config
func (c RGAConnection) Info() (*RGAResponse, error) {
	fmt.Fprintf(c, info+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// EGains returns the list of electronic gain factors available for the sensor
func (c RGAConnection) EGains() (*RGAResponse, error) {
	fmt.Fprintf(c, eGains+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], true)
}

// InletInfo returns inlet information
func (c RGAConnection) InletInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, inletInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseHorizontalResp(resp[0])
}

// RFInfo returns the current configuration and state of the RF Trip
func (c RGAConnection) RFInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, rFInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// MultiplierInfo returns the current configuration the current state of the multiplier and the reason why it is locked
func (c RGAConnection) MultiplierInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, multiplierInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// SourceInfo returns the current configuration the current state of the multiplier and the reason why it is locked
func (c RGAConnection) SourceInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, sourceInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// DetectorInfo returns a table of information about the detector settings for a particular source table
func (c RGAConnection) DetectorInfo(SourceIndex int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %d%s", detectorInfo, SourceIndex, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseHorizontalResp(resp[0])
}

// FilamentInfo returns the current config and state of the filaments
func (c RGAConnection) FilamentInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, filamentInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// TotalPressureInfo returns information about the current state and settings being used if a total pressure gauge has been fitted onto the sensor
func (c RGAConnection) TotalPressureInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, totalPressureInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// AnalogInputInfo returns information about all the analog inputs that the sensor has.
func (c RGAConnection) AnalogInputInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, analogInputInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseHorizontalResp(resp[0])
}

// AnalogOutputInfo rReturns information about all analog outputs that a sensor has
func (c RGAConnection) AnalogOutputInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, analogOutputInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseHorizontalResp(resp[0])
}

// DigitalInfo returns information about the fitted digital input ports
func (c RGAConnection) DigitalInfo() (*RGAResponse, error) { // TODO:SSSOCPaulCote this command has both vertical and horizontal outputs
	fmt.Fprintf(c, digitalInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// RolloverInfo Returns configuration settings for the rollover correction algorithm used in the HPQ2s
func (c RGAConnection) RolloverInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, rolloverInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// RVCInfo returns the current state of the RVC if the sensor has an RVC fitted.
func (c RGAConnection) RVCInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, rVCInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

//CirrusInfo returns the current Cirrus status and configuration if the sensor is a Cirrus
func (c RGAConnection) CirrusInfo() (*RGAResponse, error) {
	fmt.Fprintf(c, cirrusInfo+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

//PECal_Info It is not meant to be used by non MKS software
func (c RGAConnection) PECal_Info(SourceIndex, DetectorIndex int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %d %d%s", pECal_Info, SourceIndex, DetectorIndex, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

//Control takes an AppName (name of the TCP client) and the version of the controlling application to control an unused sensor
func (c RGAConnection) Control(AppName, Version string) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %s %s%s", control, AppName, Version, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// Release releases control of the sensor
func (c RGAConnection) Release() (*RGAResponse, error) {
	fmt.Fprintf(c, release+commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

type RGAOnOff string

const (
	Rga_ON RGAOnOff = "On"
	Rga_OFF RGAOnOff = "Off"
)

// FilamentControl turns the currently selected filament On or Off
func (c RGAConnection) FilamentControl(State RGAOnOff) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %s%s", filamentControl, State, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// FilamentSelect selects a particular filament
func (c RGAConnection) FilamentSelect(Number int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %d%s", filamentSelect, Number, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// FilamentOnTime sets the amount of time that filaments will stay on for if the unit is configured to use a time limit before filaments automatically go off
func (c RGAConnection) FilamentOnTime(Time int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %d%s", filamentOnTime, Time, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// AddAnalog adds a new analog measurement to the sensor
func (c RGAConnection) AddAnalog(Name string, StartMass, EndMass, PointerPerPeak, Accuracy, SourceIndex, DetectorIndex int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %s %d %d %d %d %d %d%s", addAnalog, Name, StartMass, EndMass, PointerPerPeak, Accuracy, SourceIndex, DetectorIndex, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

type RGAFilterMode string

const (
	Rga_PeakCenter RGAFilterMode = "PeakCenter"
	Rga_PeakMax RGAFilterMode = "PeakMax"
	Rga_PeakAverage RGAFilterMode = "PeakAverage"
)

// AddBarchart adds a new barchart measurement to the sensor
func (c RGAConnection) AddBarchart(Name string, StartMass, EndMass int, FilterMode RGAFilterMode, Accuracy, EGainIndex, SourceIndex, DetectorIndex int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %s %d %d %s %d %d %d %d%s", addBarchart, Name, StartMass, EndMass, FilterMode, Accuracy, EGainIndex, SourceIndex, DetectorIndex, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// AddPeakJump adds a new peak jump measurement to the sensor
func (c RGAConnection) AddPeakJump(Name string, FilterMode RGAFilterMode, Accuracy, EGainIndex, SourceIndex, DetectorIndex int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %s %s %d %d %d %d%s", addPeakJump, Name, FilterMode, Accuracy, EGainIndex, SourceIndex, DetectorIndex, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// AddSinglePeak adds a new single peak measurement to the sensor
func (c RGAConnection) AddSinglePeak(Name string, Mass float64, Accuracy, EGainIndex, SourceIndex, DetectorIndex int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %s %f %d %d %d %d%s", addSinglePeak, Name, Mass, Accuracy, EGainIndex, SourceIndex, DetectorIndex, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}

// MeasurementAccuracy changes the accuracy code of the currently selected measurement.
func (c RGAConnection) MeasurementAccuracy(Accuracy int) (*RGAResponse, error) {
	fmt.Fprintf(c, "%s %d%s", measurementAccuracy, Accuracy, commandSuffix)
	buf := make([]byte, BUFFER)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	resp := bytes.Split(buf, commandEnd) // The whole response minus the empty bytes leftover
	return parseVerticalResp(resp[0], false)
}