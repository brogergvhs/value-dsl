stakeholder Operator
stakeholder Rescuee
stakeholder Indirectly_Involved_Humans
stakeholder Standby_Operator
stakeholder Recipient
stakeholder Mission_Supervisor
stakeholder Ground_Rescue_Team
stakeholder Observed_Humans
stakeholder Product_Owner
stakeholder Dronology
stakeholder UAV
stakeholder Observed_Human
stakeholder FarmSupervisor

// Source Dronology0:
// If the Operator 'interaction rate with the UI drops below {{MIN_INTERACTION_THRESHOLD}} for more than 60 seconds',
// the Dronology shall display 'a Fatigue Check warning' on the 'RealTimeFlightUI'.
requirement Dronology0
if Operator interaction rate with the UI drops below MIN_INTERACTION_THRESHOLD for more than 60 seconds
system shall notify Operator using 'Fatigue Check warning on RealTimeFlightUI'
stakeholders Operator
linked_to Dronology0_source

// Source Dronology1:
// While the Operator 'measured attention level is below a predefined threshold',
// the Dronology shall turn 'the volume of audible alerts for emergency events' to '{x}dB'.
requirement Dronology1
while Operator measured attention level is below a predefined threshold
system shall alert Operator using 'audible alert volume {x}dB'
stakeholders Operator
linked_to Dronology1_source

// Source Dronology1_alt:
// While the Operator 'measured attention level is below a predefined threshold',
// the Dronology shall change 'the volume of audible alerts for emergency events' to '{x}dB'.
requirement Dronology1_alt
while Operator "measured attention level is below a predefined threshold"
system shall alert Operator using 'audible alert volume {x}dB'
stakeholders Operator
linked_to Dronology1_alt_source

// Source Dronology2:
// While the following conditions are satisfied:
// - 'a mission is active', and
// - the Operator 'biometric data indicates a stress level above {{MAX_STRESS_THRESHOLD}}'
// the Dronology shall propose 'a control handover' to a Standby_Operator.
requirement Dronology2
while Dronology mission is active
while Operator biometric data indicates a stress level above MAX_STRESS_THRESHOLD
system shall notify Standby_Operator using 'control handover proposal'
stakeholders Operator, Standby_Operator
linked_to Dronology2_source

// Source Dronology3:
// While the Operator 'has a 'Trainee' certification level',
// the Dronology shall restrict 'the execution of flight plans marked as 'Expert Level''.
requirement Dronology3
while Operator has Trainee certification level
system shall restrict 'execution of Expert Level flight plans' of Operator
stakeholders Operator
linked_to Dronology3_source

// Source Dronology4:
// If the Operator 'continuous control session exceeds the maximum duration defined by employer policy',
// the Dronology shall enforce 'a mandatory control handover procedure to a secondary operator'.
requirement Dronology4
if Operator continuous control session exceeds the maximum duration defined by employer policy
system shall notify Standby_Operator using 'mandatory control handover procedure'
stakeholders Operator, Standby_Operator
linked_to Dronology4_source

// Source Dronology5:
// When the UAV 'thermal camera detects a human heat signature matching rescue criteria',
// the Dronology shall create 'a 'Potential Target' waypoint on the map'.
requirement Dronology5
while UAV thermal camera detects a human heat signature matching rescue criteria
system shall record 'Potential Target waypoint on the map' of Operator
stakeholders Operator, Rescuee
linked_to Dronology5_source

// Source Dronology5_1:
// When the UAV 'thermal camera detects a human heat signature matching rescue criteria',
// the Dronology shall display 'a highlighted 'Potential Target' waypoint on the map' of the Operator.
requirement Dronology5_1
while UAV thermal camera detects a human heat signature matching rescue criteria
system shall notify Operator using 'highlighted Potential Target waypoint on the map'
stakeholders Operator, Rescuee
linked_to Dronology5_1_source

// Source Dronology6:
// When a Rescuee 'is located',
// the Dronology shall display 'their estimated vital signs' on the 'RealTimeFlightUI' by the means of 'sensor feeds'.
requirement Dronology6
while Rescuee is located
system shall notify Operator using 'estimated vital signs via sensor feeds'
stakeholders Rescuee, Operator
linked_to Dronology6_source

// Source Dronology7:
// While the following conditions are satisfied:
// - 'a surveillance mission is active', and
// - the UAV 'is in range of {{NO_SURVEILLANCE_ZONE_RANGE}} of a designated 'No Surveillance Zone''
// the Dronology shall ensure 'the camera is oriented such that the 'no surveillance zone' is not captured'.
requirement Dronology7
while Dronology surveillance mission is active
while UAV is in range of NO_SURVEILLANCE_ZONE_RANGE of a designated No Surveillance Zone
system shall restrict 'capture of no surveillance zone' of Indirectly_Involved_Humans
stakeholders Indirectly_Involved_Humans, Operator, Rescuee
linked_to Dronology7_source

// Source Dronology8:
// While the following conditions are satisfied:
// - 'a delivery UAV is at the final waypoint', and
// - the UAV 'camera feed visually confirms the Recipient'
// the Dronology shall allow the Operator 'to issue the 'Release Package' command'.
requirement Dronology8
while UAV delivery UAV is at the final waypoint
while UAV camera feed visually confirms the Recipient
system shall notify Operator using 'Release Package command authorization'
stakeholders Operator, Recipient
linked_to Dronology8_source

// Source Dronology9:
// If the Dronology 'detects a person not identified as the Recipient within the {{DELIVERY_SAFETY_RADIUS}} of the drop-off point',
// the Dronology shall activate 'the hover function of the UAV and alert the Operator'.
requirement Dronology9
if Dronology detects a person not identified as the Recipient within DELIVERY_SAFETY_RADIUS of the drop-off point
system shall alert Operator using 'UAV hover function'
stakeholders Operator, Recipient
linked_to Dronology9_source

// Source Dronology10:
// When the Operator 'gaze is detected averted from all mission-critical UI panels for more than {{MAX_LOOKAWAY_TIME}}',
// the Dronology shall alert the Operator 'with an audible attention prompt'.
requirement Dronology10
while Operator gaze is averted from all mission-critical UI panels for more than MAX_LOOKAWAY_TIME
system shall alert Operator using 'audible attention prompt'
stakeholders Operator
linked_to Dronology10_source

// Source Dronology11:
// While the following conditions are satisfied:
// - the Operator 'is viewing the map', and
// - the Operator 'gaze remains on a mission waypoint for more than {{MAX_LOOKAWAY_TIME}}'
// the Dronology shall display 'contextual information for that waypoint' on the 'MapComponent'.
requirement Dronology11
while Operator is viewing the map
while Operator gaze remains on a mission waypoint for more than MAX_LOOKAWAY_TIME
system shall notify Operator using 'contextual information on MapComponent'
stakeholders Operator
linked_to Dronology11_source

// Source Dronology13:
// While 'a mission is in progress',
// the Dronology shall log 'the physiological data streams' of the Operator by the means of 'timestamps corresponding to mission events'.
requirement Dronology13
while Dronology mission is in progress
system shall log 'physiological data streams' of Operator using 'timestamps corresponding to mission events'
stakeholders Operator
linked_to Dronology13_source

// Source Dronology15:
// When the UAV 'is ready for its final delivery descent',
// the Operator shall provide 'a positive visual confirmation that the landing zone is clear of bystanders' to the Dronology.
requirement Dronology15
while UAV is ready for its final delivery descent
system shall notify Dronology using 'positive visual confirmation that the landing zone is clear of bystanders'
stakeholders Operator, Indirectly_Involved_Humans, Dronology
linked_to Dronology15_source

// Source Dronology18:
// While the following conditions are satisfied:
// - 'a voice command interface is included', and
// - the Operator 'issues a valid verbal command'
// the Dronology shall perform 'the corresponding UAV operation'.
requirement Dronology18
while Dronology voice command interface is included
while Operator issues a valid verbal command
system shall notify Operator using 'corresponding UAV operation'
stakeholders Operator
linked_to Dronology18_source

// Source HMR_H01:
// While 'an active surveillance mission is ongoing' and 'multiple UAVs are involved' and the Operator 'UAV alternation rate exceeds {{MAX_CONTEXT_SWITCH_RATE}}',
// the Dronology shall propose 'assignment of the critical UAV to an available secondary operator' to the Operator.
requirement HMR_H01
while Dronology active surveillance mission is ongoing and multiple UAVs are involved
while Operator UAV alternation rate exceeds MAX_CONTEXT_SWITCH_RATE
system shall notify Operator using 'critical UAV secondary operator assignment'
stakeholders Operator, Standby_Operator
linked_to HMR_H01_source

// Source HMR_H02a:
// When 'monitoring a traffic incident and operator response time exceeds {{MAX_ALERT_RESPONSE_TIME}} for {{MAX_ALERT_CONSECUTIVE_MISSED}} consecutive alerts',
// the Dronology shall change 'alert modalities' into 'escalated state'.
requirement HMR_H02a
while Dronology monitoring a traffic incident and operator response time exceeds MAX_ALERT_RESPONSE_TIME for MAX_ALERT_CONSECUTIVE_MISSED consecutive alerts
system shall alert Mission_Supervisor using 'escalated alert modalities'
stakeholders Operator, Product_Owner, Mission_Supervisor
linked_to HMR_H02a_source

// Source HMR_H02b:
// When 'monitoring a traffic incident and operator response time exceeds {{MAX_ALERT_RESPONSE_TIME}} for {{MAX_ALERT_CONSECUTIVE_MISSED}} consecutive alerts',
// the Dronology shall notify the Mission_Supervisor.
requirement HMR_H02b
while Dronology monitoring a traffic incident and operator response time exceeds MAX_ALERT_RESPONSE_TIME for MAX_ALERT_CONSECUTIVE_MISSED consecutive alerts
system shall notify Mission_Supervisor
stakeholders Operator, Mission_Supervisor
linked_to HMR_H02b_source

// Source HMR_H03:
// While 'monitoring agricultural fields' and the UAV 'sensors detect a farm worker in close proximity to machinery',
// the Dronology shall alert the FarmSupervisor and the Operator.
requirement HMR_H03
while Dronology monitoring agricultural fields
while UAV sensors detect a farm worker in close proximity to machinery
system shall alert FarmSupervisor using 'farm worker proximity to machinery'
stakeholders Operator, Observed_Humans, Product_Owner, FarmSupervisor
linked_to HMR_H03_source

// Source HMR_H04:
// While 'UAV is operating within {{MIN_BOAT_PROXIMITY}} of a firefighter rescue boat',
// the Dronology shall require 'supervision confirmation' of the Operator every '{{x}}' Seconds.
requirement HMR_H04
while UAV operating within MIN_BOAT_PROXIMITY of a firefighter rescue boat
system shall notify Operator using 'supervision confirmation every {x} seconds'
stakeholders Operator, Product_Owner
linked_to HMR_H04_source

// Source HMR_H05a:
// If an Observed_Human 'is detected to be a human within a predefined rescue mission area',
// the Dronology shall alert the Operator.
requirement HMR_H05a
if Observed_Human is detected to be a human within a predefined rescue mission area
system shall alert Operator
stakeholders Operator, Observed_Humans
linked_to HMR_H05a_source

// Source HMR_H05b:
// If an Observed_Human 'is detected to be a human within a predefined rescue mission area',
// the Dronology shall display 'the image used for detection' to the Operator with the aim of 'confirming if the human is the rescue target'.
requirement HMR_H05b
if Observed_Human is detected to be a human within a predefined rescue mission area
system shall notify Operator using 'image used for detection'
stakeholders Operator, Observed_Humans
linked_to HMR_H05b_source

// Source HMR_H06:
// If an Observed_Human 'is detected as a rescue target',
// the Dronology shall provide 'the coordinates and contact information of the closest ground rescue team' to the Operator.
requirement HMR_H06
if Observed_Human is detected as a rescue target
system shall notify Operator using 'ground rescue team coordinates and contact information'
stakeholders Operator, Observed_Humans, Ground_Rescue_Team
linked_to HMR_H06_source
