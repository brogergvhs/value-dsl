stakeholder Driver
stakeholder Pedestrian
stakeholder Validation_Engineer
stakeholder Other_Drivers
stakeholder System
stakeholder Central_Activation_Device
stakeholder Attention_Assist
stakeholder IDC
stakeholder Engine

// Source qure114a:
// The Central_Activation_Device shall calculate 'crash acceleration and pressure forces' with the aim of 'finding threshold values' by the means of 'appropriate sensors'.
requirement qure114a
system shall measure 'crash acceleration and pressure forces' of Driver using 'appropriate sensors for finding threshold values'
stakeholders Driver, Central_Activation_Device
linked_to qure114a_source

// Source qure114b:
// The Central_Activation_Device shall determine 'if threshold values have been reached' with the aim of 'firing the seat belt tightener and/or the airbag firing squibs or for activation of the crash switches' by the means of 'qure114a'.
requirement qure114b
system shall detect 'reached threshold values' of Driver using 'qure114a for firing restraint and crash switch functions'
stakeholders Driver, Central_Activation_Device
linked_to qure114b_source

// Source qure159:
// The Central_Activation_Device shall calculate 'appropriate firing thresholds' from 'crash measurement signals under consideration of all parameters (belt buckle status, equipment programming, etc.)'.
requirement qure159
system shall measure 'appropriate firing thresholds' of Driver using 'crash measurement signals under consideration of all parameters'
stakeholders Driver, Central_Activation_Device
linked_to qure159_source

// Source qure172:
// If the System 'has a mantual transmission', the System shall detect 'driver engine stalling' with the aim of 'distinguishing between normal and bad driver actions' by the means of 'current gear and clutch information'.
requirement qure172
if System has a mantual transmission
system shall detect 'driver engine stalling' of Driver using 'current gear and clutch information'
stakeholders Driver, System
linked_to qure172_source

// Source qure226:
// The System should employ 'function specific databases (e.g., tunnels and bridges, pedestrians, ...)' with the aim of 'validating component functions'.
requirement qure226
system shall store 'function specific databases (e.g., tunnels and bridges, pedestrians, ...)' of Validation_Engineer using 'component function validation'
stakeholders Pedestrian, Validation_Engineer
linked_to qure226_source

// Source qure284a:
// The Attention_Assist shall monitor 'drowsiness' of the Driver.
requirement qure284a
system shall monitor drowsiness of Driver
stakeholders Driver, Attention_Assist
linked_to qure284a_source

// Source qure284b:
// When the Attention_Assist 'issues a drowsiness warning', the System shall activate 'an Active Comfort Program'.
requirement qure284b
while Attention_Assist issues a drowsiness warning
system shall notify Driver using 'Active Comfort Program'
stakeholders Driver, Attention_Assist
linked_to qure284b_source

// Source qure287a:
// The System shall monitor 'the area behind the vehicle without constraints'.
requirement qure287a
system shall monitor 'the area behind the vehicle without constraints' of Driver
stakeholders Driver, Pedestrian
linked_to qure287a_source

// Source qure287b:
// The System shall display 'the area behind the vehicle without constraints' to the Driver with the aim of 'parking in more easily'.
requirement qure287b
system shall notify Driver using 'area behind the vehicle without constraints for parking more easily'
stakeholders Driver, Pedestrian
linked_to qure287b_source

// Source qure378:
// If 'objects are closer than 100 m', the IDC shall provide 'objects reliably' to the System with the aim of 'ensuring synchronisation to the ADAS graphic in the IC'.
requirement qure378
if IDC objects are closer than 100 m
system shall notify System using 'reliable objects for ADAS graphic synchronisation'
stakeholders Driver, Pedestrian, IDC, System
linked_to qure378_source

// Source qure380:
// The System shall track 'pedestrian motion (e.g., motion of arms or legs)' with the aim of 'validating close passing scenarios'.
requirement qure380
system shall track 'pedestrian motion (e.g., motion of arms or legs)' of Pedestrian using 'close passing scenario validation'
stakeholders Validation_Engineer, Pedestrian
linked_to qure380_source

// Source qure462:
// The System shall monitor 'the temperature of the vehicle inlet close to the DC pins' by the means of 'a second temperature sensor (type PT1000) integrated into the vehicle inlet'.
requirement qure462
system shall monitor 'the temperature of the vehicle inlet close to the DC pins' of Driver using 'a second temperature sensor (type PT1000) integrated into the vehicle inlet'
stakeholders Driver
linked_to qure462_source

// Source qure469a:
// The System shall track 'the distance' of a Pedestrian.
requirement qure469a
system shall track 'the distance' of Pedestrian
stakeholders Pedestrian
linked_to qure469a_source

// Source qure469b:
// When a Pedestrian 'projected movement will bring them into the pedestrian warning area (TSA5_Func-319) at the time of the EV reaching the crosswalk', the System shall classify 'a pedestrian' as 'close to the crosswalk'.
requirement qure469b
while Pedestrian projected movement will bring them into the pedestrian warning area (TSA5_Func-319) at the time of the EV reaching the crosswalk
system shall detect 'pedestrian close to the crosswalk' of Pedestrian
stakeholders Pedestrian
linked_to qure469b_source

// Source qure497a:
// When the System 'is close at a crosswalk', the System shall monitor 'the location around the car' with the aim of 'detecting parellel driving cyclists'.
requirement qure497a
while System is close at a crosswalk
system shall monitor 'the location around the car' of Pedestrian using 'detecting parellel driving cyclists'
stakeholders Pedestrian, System
linked_to qure497a_source

// Source qure497b:
// If 'the only objects close to a crosswalk are cyclists driving parallel to the System's direction of travel', the System shall not alert the Driver of 'crosswalk warnings'.
requirement qure497b
if System "the only objects close to a crosswalk are cyclists driving parallel to the System's direction of travel"
system shall limit 'crosswalk warnings' of Driver
stakeholders Pedestrian, Driver, System
linked_to qure497b_source

// Source qure500a:
// The System shall detect 'close objects' by the means of 'ultrasonic sensors'.
requirement qure500a
system shall detect 'close objects' of Driver using 'ultrasonic sensors'
stakeholders Driver, Pedestrian
linked_to qure500a_source

// Source qure500b:
// The System shall display 'discrete distance information of close objects' to the Driver.
requirement qure500b
system shall notify Driver using 'discrete distance information of close objects'
stakeholders Driver, Pedestrian
linked_to qure500b_source

// Source qure501:
// The System shall assist the Driver with 'maneuvering the vehicle close to obstacles in front of the vehicle'.
requirement qure501
system shall notify Driver using 'maneuvering the vehicle close to obstacles in front of the vehicle'
stakeholders Driver
linked_to qure501_source

// Source qure915:
// The System shall detect 'if an object does not fulfill the requirements for a pedestrian (incl. pedestrian dummies, pedestrians with sports equipment) independant of their moving state (moving, stationary)' with the aim of 'enabling fast functional reaction after camera fusion'.
requirement qure915
system shall detect 'object does not fulfill pedestrian requirements independent of moving state' of Pedestrian using 'fast functional reaction after camera fusion'
stakeholders Driver, Pedestrian
linked_to qure915_source

// Source qure922a:
// The System shall track 'brake paddle pressing thresholds'.
requirement qure922a
system shall track 'brake paddle pressing thresholds' of Driver
stakeholders Driver
linked_to qure922a_source

// Source qure922b:
// If 'the driver touches the brake pedal fast but not deep enough', the System shall activate 'additional braking beyond the dirver's input'.
requirement qure922b
if Driver touches the brake pedal fast but not deep enough
system shall notify Driver using "additional braking beyond the dirver's input"
stakeholders Driver
linked_to qure922b_source

// Source qure1050a:
// The System shall detect 'Pedestrians in front of the System' by the means of 'long range radar'.
requirement qure1050a
system shall detect 'Pedestrians in front of the System' of Pedestrian using 'long range radar'
stakeholders Pedestrian
linked_to qure1050a_source

// Source qure1050b:
// The System shall calculate 'position (x, y), velocity (vx, vy), and quality measures' from 'detected Pedestrians'.
requirement qure1050b
system shall measure 'position (x, y), velocity (vx, vy), and quality measures' of Pedestrian using 'detected Pedestrians'
stakeholders Pedestrian
linked_to qure1050b_source

// Source qure1178a:
// The System shall record 'the belt buckle status'.
requirement qure1178a
system shall record 'the belt buckle status' of Driver
stakeholders Driver
linked_to qure1178a_source

// Source qure1178b:
// While 'the belt buckle status cannot be recorded', the System shall display 'Signal Not Available message' on the 'CAN'.
requirement qure1178b
while System the belt buckle status cannot be recorded
system shall notify Driver using 'Signal Not Available message on CAN'
stakeholders Driver, System
linked_to qure1178b_source

// Source qure1431a:
// The System shall detect 'objects on side tracks and in front of the preceding vehicle'.
requirement qure1431a
system shall detect 'objects on side tracks and in front of the preceding vehicle' of Driver
stakeholders Driver, Other_Drivers, Pedestrian
linked_to qure1431a_source

// Source qure1431b:
// The System shall calculate 'objects on side tracks and in front of the preceding vehicle'.
requirement qure1431b
system shall measure 'objects on side tracks and in front of the preceding vehicle' of Driver
stakeholders Driver, Other_Drivers, Pedestrian
linked_to qure1431b_source

// Source qure1876a:
// The System shall track 'the presence' of a Driver.
requirement qure1876a
system shall track 'the presence' of Driver
stakeholders Driver
linked_to qure1876a_source

// Source qure1876b:
// While the Engine 'is in automatic stop state' and 'absence of the driver is detected', the System shall require 'a level 1 restart' for the Engine.
requirement qure1876b
while Engine is in automatic stop state and absence of the driver is detected
system shall notify Driver using 'level 1 restart for the Engine'
stakeholders Driver, Engine
linked_to qure1876b_source
