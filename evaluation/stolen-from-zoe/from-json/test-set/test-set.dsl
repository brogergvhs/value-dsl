stakeholder Patient
stakeholder Caregiver
stakeholder Relative
stakeholder Doctor
stakeholder Driver
stakeholder Pedestrian
stakeholder System_Administrator
stakeholder Meeting_Participant
stakeholder Vendor
stakeholder GMCB_Staff
stakeholder User
stakeholder System
stakeholder Meeting_Scheduler_System

// Source ACTIVAGE0:
// The System shall measure 'the commitment to the exercise routine' of the Patient.
requirement ACTIVAGE0
system shall measure 'the commitment to the exercise routine' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE0_source

// Source ACTIVAGE1:
// The System shall record 'blood pressure' of the Patient.
requirement ACTIVAGE1
system shall record 'blood pressure' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE1_source

// Source ACTIVAGE2:
// The System shall monitor 'the social activities (outdoor activities, etc.)' of the Patient.
requirement ACTIVAGE2
system shall monitor 'the social activities (outdoor activities, etc.)' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE2_source

// Source ACTIVAGE3_1:
// The System shall monitor 'factors relating to falls, gait, speed etc.' of the Patient.
requirement ACTIVAGE3_1
system shall monitor 'factors relating to falls, gait, speed etc.' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE3_1_source

// Source ACTIVAGE4:
// The System shall detect 'abnormal activities and unusual situation'.
requirement ACTIVAGE4
system shall detect 'abnormal activities and unusual situation' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE4_source

// Source ACTIVAGE5:
// The System shall obtain 'all information related to a car (position, speed, direction)'.
requirement ACTIVAGE5
system shall collect 'all information related to a car (position, speed, direction)' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE5_source

// Source ACTIVAGE6:
// The System shall alert the Caregiver and Relative of 'an anomaly occurring' of the Patient every 'single' night.
requirement ACTIVAGE6
while Patient anomaly occurring every single night
system shall alert Caregiver using 'patient anomaly notification'
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE6_source

// Source ACTIVAGE7:
// The System shall track 'the number of steps' of the Patient every 'single' day.
requirement ACTIVAGE7
while Patient every single day
system shall track 'the number of steps' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE7_source

// Source ACTIVAGE8:
// The System shall collect 'information about which users are attending some events/activities' at the 'home' of the Patient.
requirement ACTIVAGE8
where at the home of the Patient
system shall collect 'information about which users are attending some events/activities' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE8_source

// Source ACTIVAGE9:
// The System shall record 'weight' of the Patient.
requirement ACTIVAGE9
system shall record weight of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE9_source

// Source ACTIVAGE10:
// The System shall record 'speech' of the Patient.
requirement ACTIVAGE10
system shall record speech of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE10_source

// Source ACTIVAGE11:
// The System shall include 'an air quality control device'.
requirement ACTIVAGE11
system shall record 'an air quality control device' of System
stakeholders Patient, Caregiver, Relative, Doctor, System
linked_to ACTIVAGE11_source

// Source ACTIVAGE12_1:
// The System shall record 'the amount of water drank' of the Patient.
requirement ACTIVAGE12_1
system shall record 'the amount of water drank' of Patient
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE12_1_source

// Source ACTIVAGE12_2:
// If the Patient 'does not drink enough water', the System shall inform the Patient to 'drink more water'.
requirement ACTIVAGE12_2
if Patient does not drink enough water
system shall inform Patient using 'drink more water'
stakeholders Patient, Caregiver, Relative, Doctor
linked_to ACTIVAGE12_2_source

// Source MobSTr0:
// The System shall detect 'all obstacles'.
requirement MobSTr0
system shall detect 'all obstacles' of Driver
stakeholders Driver
linked_to MobSTr0_source

// Source MobSTr1:
// The System shall identify 'pedestrians, cars, trucks, busses, motorbikes, bicycles, riders, traffic lights, traffic signs'.
requirement MobSTr1
system shall detect 'pedestrians, cars, trucks, busses, motorbikes, bicycles, riders, traffic lights, traffic signs' of Driver
stakeholders Driver, Pedestrian
linked_to MobSTr1_source

// Source MobSTr0_1:
// The System shall detect 'all obstacles' by the means of 'more than one sensor'.
requirement MobSTr0_1
system shall detect 'all obstacles' of Driver using 'more than one sensor'
stakeholders Driver
linked_to MobSTr0_1_source

// Source PROMISE0:
// The System shall provide 'GUI based Monitoring Services' to a System_Administrator with the aim of 'allowing the System_Administrator to monitor message exchanges'.
requirement PROMISE0
system shall notify System_Administrator using 'GUI based Monitoring Services for message exchange monitoring'
stakeholders System_Administrator
linked_to PROMISE0_source

// Source PROMISE1:
// If 'no location is available or feasible for the meeting',
// the Meeting_Scheduler_System shall propose 'a virtual meeting for a given time slot' to the Meeting_Participant.
requirement PROMISE1
if Meeting_Scheduler_System no location is available or feasible for the meeting
system shall notify Meeting_Participant using 'virtual meeting for a given time slot'
stakeholders Meeting_Participant, Meeting_Scheduler_System
linked_to PROMISE1_source

// Source VHCURES0_1:
// The Vendor shall provide 'status reports about data submission and conformance to GMCBs specifications' to the GMCB_Staff at a minimum of every 'single' month with the aim of 'supporting GMCB oversight of non-compliant data submitters'.
requirement VHCURES0_1
while Vendor at a minimum of every single month
system shall notify GMCB_Staff using 'status reports about data submission and conformance to GMCBs specifications supporting oversight'
stakeholders GMCB_Staff, Vendor
linked_to VHCURES0_1_source

// Source VHCURES0_2:
// The Vendor shall provide 'status reports about data submission and conformance to GMCBs specifications' to the GMCB_Staff upon request.
requirement VHCURES0_2
while GMCB_Staff upon request
system shall notify GMCB_Staff using 'status reports about data submission and conformance to GMCBs specifications'
stakeholders GMCB_Staff, Vendor
linked_to VHCURES0_2_source

// Source VHCURES0_3:
// The Vendor shall provide 'sample of what fields/content types that would be provided as part of these reports' to the GMCB_Staff.
requirement VHCURES0_3
system shall notify GMCB_Staff using 'sample of fields and content types provided as part of these reports'
stakeholders GMCB_Staff, Vendor
linked_to VHCURES0_3_source

// Source VHCURES2:
// The Vendor shall provide 'status reports about data submission and conformance to specifications' to the GMCB_Staff every 'single' month with the aim of 'supporting GMCB oversight of non-compliant data submitters'.
requirement VHCURES2
while Vendor every single month
system shall notify GMCB_Staff using 'status reports about data submission and conformance to specifications supporting oversight'
stakeholders GMCB_Staff, Vendor
linked_to VHCURES2_source

// Source VHCURES3:
// The System shall log 'unsuccessful login attempts'.
requirement VHCURES3
system shall log 'unsuccessful login attempts' of User
stakeholders Vendor, GMCB_Staff, User
linked_to VHCURES3_source

// Source WHO0:
// The System shall log 'system logins and logouts'.
requirement WHO0
system shall log 'system logins and logouts' of User
stakeholders User
linked_to WHO0_source

// Source WHO1:
// The System shall record 'all authentication violations'.
requirement WHO1
system shall record 'all authentication violations' of User
stakeholders User
linked_to WHO1_source
