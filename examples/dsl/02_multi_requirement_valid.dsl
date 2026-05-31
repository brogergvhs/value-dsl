stakeholder Worker
stakeholder Supervisor
stakeholder SafetyOfficer

requirement R1
when Worker enters RestrictedZone
system shall track location of Worker using Camera
stakeholders Worker, Supervisor

requirement R2
when Worker enters HazardousMaterialZone
system shall monitor protective_equipment_usage of Worker using Sensor
stakeholders Worker, SafetyOfficer

requirement R3
when Worker enters LoadingDock
system shall monitor speed_and_route of Worker using Tracker
stakeholders Worker, Supervisor
