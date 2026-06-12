// Generated from evaluation_requirements_human(1).csv and mapped to value-dsl
// Scope restriction: stakeholder declarations + requirement definitions only
// No value declarations and no assignment declarations are used.

stakeholder ambulatory_care_system, application_system, appropriate_sensors, asd, attention_assist, bedside_monitor, camera_fusion, can_bus
stakeholder caregiver, central_activation_device, classroom_manager, component, connected_vehicle_system, conveyance, crash_measurement_unit, cyclist
stakeholder data_link, driver, ehr_system, engine_controller, exposure_source, hci, host_vehicle, icu_component
stakeholder idc, investigator, location, long_range_radar, maintainer, nurse, nyc_cvpd_subsystem, om_system
stakeholder operator, organization, patient, payload, pedestrian, person, physician, physiologic_monitor
stakeholder public_gathering, remote_location, sample, second_temperature_sensor, shipper, site, software, tcs
stakeholder telemetry_system, threshold_calculator, transit_vehicle_operator, trip, ultrasonic_sensors, user, vehicle, vehicle_inlet

// Row 001 | source_id: 114 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_001_ID_X_114
  system shall measure threshold_status of central_activation_device
  stakeholders central_activation_device, appropriate_sensors, threshold_calculator
  linked_to "114"

// Row 002 | source_id: 159 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_002_ID_X_159
  system shall measure belt_buckle_status of crash_measurement_unit
  stakeholders crash_measurement_unit, threshold_calculator
  linked_to "159"

// Row 003 | source_id: 172 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_003_ID_X_172
  where manual_transmission_car
  system shall inform driver
  stakeholders software, driver, vehicle, engine_controller
  linked_to "172"

// Row 004 | source_id: 226 | definability: no
// Issue: Statement is a process/background/inventory requirement rather than a stakeholder-
// centered requirement definition supported by Value-DSL.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: For validation of certain component functions a function specific
// data base may be necessary, e.g. tunnels and bridges, pedestrians, oncoming or crossing objects,
// situations with high yaw dynamic, lane radius verification, etc.

// Row 005 | source_id: 227 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_005_ID_X_227
  system shall detect detection_done_in_certain of application_system
  stakeholders application_system
  linked_to "227"

// Row 006 | source_id: 284 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_006_ID_X_284
  when attention_assist reports drowsiness_warning
  system shall warn om_system
  stakeholders attention_assist, om_system
  linked_to "284"

// Row 007 | source_id: 287 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Compound requirement was collapsed to one primary
// Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_007_ID_X_287
  system shall inform driver using display
  stakeholders driver, vehicle
  linked_to "287"

// Row 008 | source_id: 288 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_008_ID_X_288
  system shall inform application_system using display
  stakeholders application_system
  linked_to "288"

// Row 009 | source_id: 378 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Quantitative/timing/capacity constraints are not
// first-class in Value-DSL and were preserved only implicitly.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_009_ID_X_378
  where condition_from_source_9
  system shall inform idc
  stakeholders idc
  linked_to "378"

// Row 010 | source_id: 380 | definability: no
// Issue: Statement is a process/background/inventory requirement rather than a stakeholder-
// centered requirement definition supported by Value-DSL.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: close passing of real PD for validation of motion (e.g. motion of
// arms or legs, etc.)

// Row 011 | source_id: 462 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_011_ID_X_462
  system shall monitor temperature of appropriate_sensors
  stakeholders appropriate_sensors, vehicle_inlet, second_temperature_sensor, vehicle
  linked_to "462"

// Row 012 | source_id: 469 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_012_ID_X_469
  where condition_from_source_12
  system shall warn pedestrian
  stakeholders pedestrian
  linked_to "469"

// Row 013 | source_id: 497 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_013_ID_X_497
  where condition_from_source_13
  system shall limit crosswalk_warnings of cyclist
  stakeholders cyclist, vehicle, trip
  linked_to "497"

// Row 014 | source_id: 498 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_014_ID_X_498
  where condition_from_source_14
  system shall warn driver
  stakeholders driver, pedestrian, vehicle
  linked_to "498"

// Row 015 | source_id: 499 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_015_ID_X_499
  system shall inform vehicle using display
  stakeholders vehicle
  linked_to "499"

// Row 016 | source_id: 500 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_016_ID_X_500
  system shall inform ultrasonic_sensors using display
  stakeholders appropriate_sensors, ultrasonic_sensors
  linked_to "500"

// Row 017 | source_id: 501 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_017_ID_X_501
  system shall inform driver using display
  stakeholders driver, vehicle, om_system
  linked_to "501"

// Row 018 | source_id: 892 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_018_ID_X_892
  system shall detect threshold_status of threshold_calculator
  stakeholders threshold_calculator, pedestrian, camera_fusion
  linked_to "892"

// Row 019 | source_id: 915 | definability: no
// Issue: No supported Value-DSL requirement action could be identified without inventing
// semantics.
// Note: Kept as comment only.
// Unmapped source requirement: The signal has to signalize (signal state 0 equal NACPT) that an
// object does not fulfill the requirements for a pedestrian (incl. pedestrian dummies, pedestrians
// with sports equipment or other equipment) independently of their moving state (moving,
// stationary) for a fast functional reaction exclusively after a fusion with a camera, i.

// Row 020 | source_id: 922 | definability: no
// Issue: No supported Value-DSL requirement action could be identified without inventing
// semantics.
// Note: Kept as comment only.
// Unmapped source requirement: The standard Brake-Assistant is a system which brakes more than the
// driver, if the driver touches the brake pedal fast but not deep enough.

// Row 021 | source_id: 1050 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_021_ID_X_1050
  system shall inform long_range_radar
  stakeholders vehicle, long_range_radar
  linked_to "1050"

// Row 022 | source_id: 1178 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_022_ID_X_1178
  system shall inform can_bus
  stakeholders can_bus
  linked_to "1178"

// Row 023 | source_id: 1430 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_023_ID_X_1430
  system shall detect targets of component
  stakeholders component, om_system
  linked_to "1430"

// Row 024 | source_id: 1431 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb. Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_024_ID_X_1431
  system shall track objects of vehicle
  stakeholders vehicle, component, om_system
  linked_to "1431"

// Row 025 | source_id: 1876 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_025_ID_X_1876
  if engine_controller engine_condition_requires_action
  system shall inform driver
  stakeholders driver, engine_controller, om_system
  linked_to "1876"

// Row 026 | source_id: G.D9.3.0.5 | definability: no
// Issue: Statement is a process/background/inventory requirement rather than a stakeholder-
// centered requirement definition supported by Value-DSL.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: The operational capabilities to be performed by the system will be
// determined by task analysis in accordance with MIL STD 1388 Task 401 as a guide based on a
// thorough understanding of Outrider and Predator mission requirements. Tasks will be evaluated
// and allocated based on operator skills and proficiencies. The initial TCS task analysis will
// produce a system baseline which will be optimized by engineering analysis and operator
// evaluations.

// Row 027 | source_id: G.D9.3.1.1.0.8 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_027_ID_G_D9_3_1_1_0_8
  where condition_from_source_27
  system shall inform operator
  stakeholders tcs, operator
  linked_to "G.D9.3.1.1.0.8"

// Row 028 | source_id: G.D9.3.1.1.2.0.3 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_028_ID_G_D9_3_1_1_2_0_3
  system shall inform om_system
  stakeholders tcs, om_system
  linked_to "G.D9.3.1.1.2.0.3"

// Row 029 | source_id: G.D9.3.2.1.1.0.5 | definability: no
// Issue: Compound feature inventory with many sub-capabilities; no single valid Value-DSL
// stakeholder/action definition without splitting and inventing semantics.
// Note: Kept as comment only.
// Unmapped source requirement: The TCS flight route planner shall include, as a minimum, the
// following flight planning tools: 1. Weight and balance take off data calculations. 2. Fuel
// Calculations. 3. Terrain avoidance warning for line of sight flights. 4. Minimum data link
// reception altitude calculations for line of sight flights. 5. Payload search area information
// such as: visual acuity range due to atmospheric conditions, diurnal transition periods for
// thermal imagery, and lunar and solar terrain shadowing. 6. Ability to designate flight corridors
// and restricted air space.

// Row 030 | source_id: G.D9.3.2.1.1.1.0.2 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_030_ID_G_D9_3_2_1_1_1_0_2
  system shall inform operator using display
  stakeholders can_bus, tcs, operator
  linked_to "G.D9.3.2.1.1.1.0.2"

// Row 031 | source_id: G.D9.3.2.2.1.5.0.1 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_031_ID_G_D9_3_2_2_1_5_0_1
  system shall inform telemetry_system using display
  stakeholders tcs, telemetry_system
  linked_to "G.D9.3.2.2.1.5.0.1"

// Row 032 | source_id: G.D9.3.2.2.2.2.1.1 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Compound requirement was collapsed to one primary
// Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_032_ID_G_D9_3_2_2_2_2_1_1
  system shall inform payload using display
  stakeholders tcs, payload
  linked_to "G.D9.3.2.2.2.2.1.1"

// Row 033 | source_id: G.D9.3.2.2.3.0.2 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_033_ID_G_D9_3_2_2_3_0_2
  system shall record tcs_functionality_necessary_record of tcs
  stakeholders tcs, data_link
  linked_to "G.D9.3.2.2.3.0.2"

// Row 034 | source_id: G.D9.3.2.2.3.1.0.4 | definability: no
// Issue: Statement is a process/background/inventory requirement rather than a stakeholder-
// centered requirement definition supported by Value-DSL.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: As a minimum the TCS LOS data terminal control modes shall include
// acquisition, autotrack, search, manual point, omni directional, as well as directional modes of
// operation, if applicable to the selected data link.

// Row 035 | source_id: G.D9.3.2.3.1.1 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Quantitative/timing/capacity constraints are not
// first-class in Value-DSL and were preserved only implicitly. Compound requirement was collapsed
// to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_035_ID_G_D9_3_2_3_1_1
  system shall inform om_system using display
  stakeholders long_range_radar, tcs, payload, telemetry_system, om_system
  linked_to "G.D9.3.2.3.1.1"

// Row 036 | source_id: G.D9.3.2.3.1.2 | definability: partially
// Issue: Quantitative/timing/capacity constraints are not first-class in Value-DSL and were
// preserved only implicitly.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_036_ID_G_D9_3_2_3_1_2
  system shall store tcs_able_store_up of can_bus
  stakeholders can_bus, tcs, payload
  linked_to "G.D9.3.2.3.1.2"

// Row 037 | source_id: G.D9.3.2.3.2.1 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_037_ID_G_D9_3_2_3_2_1
  system shall inform operator using display
  stakeholders tcs, operator, om_system
  linked_to "G.D9.3.2.3.2.1"

// Row 038 | source_id: G.D9.3.2.3.2.5 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Compound requirement was collapsed to one primary
// Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_038_ID_G_D9_3_2_3_2_5
  system shall inform operator using display
  stakeholders tcs, operator
  linked_to "G.D9.3.2.3.2.5"

// Row 039 | source_id: G.D9.3.2.3.3.2 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_039_ID_G_D9_3_2_3_3_2
  system shall inform tcs using display
  stakeholders tcs
  linked_to "G.D9.3.2.3.3.2"

// Row 040 | source_id: G.D9.3.2.3.3.3 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Quantitative/timing/capacity constraints are not
// first-class in Value-DSL and were preserved only implicitly. Compound requirement was collapsed
// to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_040_ID_G_D9_3_2_3_3_3
  where condition_from_source_40
  system shall inform location using display
  stakeholders vehicle, tcs, location
  linked_to "G.D9.3.2.3.3.3"

// Row 041 | source_id: G.D9.3.2.4.0.2 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_041_ID_G_D9_3_2_4_0_2
  system shall inform operator using display
  stakeholders can_bus, tcs, operator, payload, location
  linked_to "G.D9.3.2.4.0.2"

// Row 042 | source_id: G.D9.3.2.4.1.1 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_042_ID_G_D9_3_2_4_1_1
  system shall inform location using display
  stakeholders tcs, payload, location
  linked_to "G.D9.3.2.4.1.1"

// Row 043 | source_id: G.D9.3.10.1.0.1 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_043_ID_G_D9_3_10_1_0_1
  system shall capture telemetry_elements of vehicle
  stakeholders vehicle, tcs, payload, data_link, telemetry_system, om_system
  linked_to "G.D9.3.10.1.0.1"

// Row 044 | source_id: G.D9.3.10.3.0.1 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_044_ID_G_D9_3_10_3_0_1
  system shall capture telemetry_elements of software
  stakeholders software, vehicle, tcs, payload, data_link, telemetry_system, om_system
  linked_to "G.D9.3.10.3.0.1"

// Row 045 | source_id: G.D9.3.13.1.9 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_045_ID_G_D9_3_13_1_9
  system shall warn operator
  stakeholders tcs, operator, om_system
  linked_to "G.D9.3.13.1.9"

// Row 046 | source_id: G.D9.3.13.1.12 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_046_ID_G_D9_3_13_1_12
  system shall warn hci using display
  stakeholders tcs, hci
  linked_to "G.D9.3.13.1.12"

// Row 047 | source_id: G.D9.3.13.1.13 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_047_ID_G_D9_3_13_1_13
  system shall inform operator
  stakeholders tcs, operator, hci, om_system
  linked_to "G.D9.3.13.1.13"

// Row 048 | source_id: G.D9.3.13.1.15 | definability: no
// Issue: Statement is a process/background/inventory requirement rather than a stakeholder-
// centered requirement definition supported by Value-DSL.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: The operational tasks to be performed concurrently by the operator
// during normal operation will be determined by appropriate task analysis and function allocation.

// Row 049 | source_id: G.D9.3.13.1.29 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Compound requirement was collapsed to one primary
// Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_049_ID_G_D9_3_13_1_29
  system shall inform operator using display
  stakeholders tcs, operator
  linked_to "G.D9.3.13.1.29"

// Row 050 | source_id: G.D9.3.14.9 | definability: no
// Issue: Statement is a process/background/inventory requirement rather than a stakeholder-
// centered requirement definition supported by Value-DSL.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: Training shall be adequate to maintain operator and maintainer
// skills and proficiencies.

// Row 051 | source_id: G.D9.3.14.10 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_051_ID_G_D9_3_14_10
  system shall record operator_maintainer_actions of tcs
  stakeholders tcs, operator, maintainer
  linked_to "G.D9.3.14.10"

// Row 052 | source_id: G.D9.3.14.11 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_052_ID_G_D9_3_14_11
  system shall detect performance_parameters of tcs
  stakeholders tcs, operator, maintainer, om_system
  linked_to "G.D9.3.14.11"

// Row 053 | source_id: G.D32.6.31 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Compound requirement was collapsed to one primary
// Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_053_ID_G_D32_6_31
  where condition_from_source_53
  system shall inform om_system
  stakeholders classroom_manager, om_system
  linked_to "G.D32.6.31"

// Row 054 | source_id: G.D33.2.0.2 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_054_ID_G_D33_2_0_2
  system shall collect demographic_data of om_system
  stakeholders om_system, person, trip, conveyance, sample
  linked_to "G.D33.2.0.2"

// Row 055 | source_id: G.D33.2.0.3 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_055_ID_G_D33_2_0_3
  system shall monitor systems_supporting_om_case of om_system
  stakeholders om_system
  linked_to "G.D33.2.0.3"

// Row 056 | source_id: G.D33.2.2.1.1 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_056_ID_G_D33_2_2_1_1
  system shall capture demographic_data of om_system
  stakeholders om_system, person
  linked_to "G.D33.2.2.1.1"

// Row 057 | source_id: G.D33.2.2.1.2 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_057_ID_G_D33_2_2_1_2
  system shall inform location
  stakeholders om_system, organization, location
  linked_to "G.D33.2.2.1.2"

// Row 058 | source_id: G.D33.2.2.1.3 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_058_ID_G_D33_2_2_1_3
  system shall store location_data of om_system
  stakeholders om_system, location
  linked_to "G.D33.2.2.1.3"

// Row 059 | source_id: G.D33.2.2.1.7 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_059_ID_G_D33_2_2_1_7
  system shall capture location_data of om_system
  stakeholders om_system, location, public_gathering
  linked_to "G.D33.2.2.1.7"

// Row 060 | source_id: G.D33.2.2.1.9 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_060_ID_G_D33_2_2_1_9
  system shall capture travel_history of om_system
  stakeholders om_system, trip
  linked_to "G.D33.2.2.1.9"

// Row 061 | source_id: G.D33.2.2.3.1.1 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_061_ID_G_D33_2_2_3_1_1
  where condition_from_source_61
  system shall inform conveyance
  stakeholders om_system, person, trip, conveyance
  linked_to "G.D33.2.2.3.1.1"

// Row 062 | source_id: G.D33.2.2.3.3 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_062_ID_G_D33_2_2_3_3
  where condition_from_source_62
  system shall inform conveyance
  stakeholders vehicle, om_system, conveyance
  linked_to "G.D33.2.2.3.3"

// Row 063 | source_id: G.D33.2.2.4.0 | definability: no
// Issue: Statement is a process/background/inventory requirement rather than a stakeholder-
// centered requirement definition supported by Value-DSL.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: Case Investigation and Exposure Contact Data Case and exposure data
// provide more detailed information beyond demographic data. Cases can be persons or animals, and
// exposure contacts can be persons, animals, other organisms, or exposure settings , such as
// travel conveyance, location, organization, object, or event.

// Row 064 | source_id: G.D33.2.2.4.2.3 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_064_ID_G_D33_2_2_4_2_3
  system shall log epi_data of om_system
  stakeholders om_system
  linked_to "G.D33.2.2.4.2.3"

// Row 065 | source_id: G.D33.2.2.4.2.5 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_065_ID_G_D33_2_2_4_2_5
  system shall record exposure_linkages of om_system
  stakeholders om_system
  linked_to "G.D33.2.2.4.2.5"

// Row 066 | source_id: G.D33.2.2.4.2.6 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_066_ID_G_D33_2_2_4_2_6
  system shall inform investigator
  stakeholders person, investigator
  linked_to "G.D33.2.2.4.2.6"

// Row 067 | source_id: G.D33.2.2.4.3.2 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_067_ID_G_D33_2_2_4_3_2
  system shall collect exposure_source_data of om_system
  stakeholders om_system, person, exposure_source
  linked_to "G.D33.2.2.4.3.2"

// Row 068 | source_id: G.D33.2.2.4.3.3 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_068_ID_G_D33_2_2_4_3_3
  system shall collect epi_data of om_system
  stakeholders om_system
  linked_to "G.D33.2.2.4.3.3"

// Row 069 | source_id: G.D33.2.2.4.3.4 | definability: partially
// Issue: Generic capability/support wording was approximated as a record requirement; Value-DSL
// lacks native capability/provide/support constructs.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_069_ID_G_D33_2_2_4_3_4
  system shall record systems_supporting_om_capturing of om_system
  stakeholders om_system, person
  linked_to "G.D33.2.2.4.3.4"

// Row 070 | source_id: G.D33.2.2.5.1 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_070_ID_G_D33_2_2_5_1
  where condition_from_source_70
  system shall monitor systems_supporting_om_monitoring of om_system
  stakeholders om_system
  linked_to "G.D33.2.2.5.1"

// Row 071 | source_id: G.D33.2.2.5.1.1 | definability: partially
// Issue: Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_071_ID_G_D33_2_2_5_1_1
  system shall collect monitoring_data of om_system
  stakeholders om_system
  linked_to "G.D33.2.2.5.1.1"

// Row 072 | source_id: G.D33.2.2.5.1.2 | definability: partially
// Issue: Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_072_ID_G_D33_2_2_5_1_2
  system shall collect follow_up_data of om_system
  stakeholders om_system, person
  linked_to "G.D33.2.2.5.1.2"

// Row 073 | source_id: G.D33.2.2.6.5 | definability: partially
// Issue: Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_073_ID_G_D33_2_2_6_5
  system shall inform shipper
  stakeholders person, location, sample, shipper
  linked_to "G.D33.2.2.6.5"

// Row 074 | source_id: G.D33.2.2.7.2 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_074_ID_G_D33_2_2_7_2
  system shall inform patient
  stakeholders om_system, patient
  linked_to "G.D33.2.2.7.2"

// Row 075 | source_id: G.D33.2.2.8.1 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_075_ID_G_D33_2_2_8_1
  where condition_from_source_75
  system shall collect adverse_event_data of person
  stakeholders person
  linked_to "G.D33.2.2.8.1"

// Row 076 | source_id: G.D33.2.2.9.1 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_076_ID_G_D33_2_2_9_1
  system shall inform om_system
  stakeholders om_system
  linked_to "G.D33.2.2.9.1"

// Row 077 | source_id: G.D33.2.2.9.2 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_077_ID_G_D33_2_2_9_2
  system shall log activity_logs of om_system
  stakeholders om_system, investigator
  linked_to "G.D33.2.2.9.2"

// Row 078 | source_id: G.D33.2.3.2.4 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_078_ID_G_D33_2_3_2_4
  system shall inform om_system
  stakeholders om_system
  linked_to "G.D33.2.3.2.4"

// Row 079 | source_id: G.D33.2.3.3.2 | definability: partially
// Issue: Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_079_ID_G_D33_2_3_3_2
  system shall monitor exposed_entity_contacts of om_system
  stakeholders om_system
  linked_to "G.D33.2.3.3.2"

// Row 080 | source_id: 53 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_080_ID_X_53
  system shall detect authorization_status of om_system
  stakeholders om_system, ehr_system
  linked_to "53"

// Row 081 | source_id: 101 | definability: partially
// Issue: Audit semantics were mapped to logging because Value-DSL has no dedicated audit verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_081_ID_X_101
  system shall log location_data of location
  stakeholders location, ehr_system, site, remote_location
  linked_to "101"

// Row 082 | source_id: 114 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_082_ID_X_114
  system shall warn ehr_system using display
  stakeholders person, patient, ehr_system
  linked_to "114"

// Row 083 | source_id: 115 | definability: partially
// Issue: Generic capability/support wording was approximated as a record requirement; Value-DSL
// lacks native capability/provide/support constructs.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_083_ID_X_115
  system shall record order_safety_checks of can_bus
  stakeholders can_bus, ehr_system, user
  linked_to "115"

// Row 084 | source_id: 123 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_084_ID_X_123
  system shall capture encounter_data of vehicle
  stakeholders vehicle, ehr_system, ambulatory_care_system
  linked_to "123"

// Row 085 | source_id: 126 | definability: partially
// Issue: Generic capability/support wording was approximated as a record requirement; Value-DSL
// lacks native capability/provide/support constructs. Compound requirement was collapsed to one
// primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_085_ID_X_126
  system shall record include_report_construction_feature of can_bus
  stakeholders can_bus, patient, ehr_system, site
  linked_to "126"

// Row 086 | source_id: 128 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_086_ID_X_128
  system shall warn ehr_system
  stakeholders vehicle, ehr_system
  linked_to "128"

// Row 087 | source_id: 137 | definability: partially
// Issue: Generic capability/support wording was approximated as a record requirement; Value-DSL
// lacks native capability/provide/support constructs.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_087_ID_X_137
  system shall record include_non_standard_laboratory of ehr_system
  stakeholders ehr_system
  linked_to "137"

// Row 088 | source_id: G.D77.3.2.1.3 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_088_ID_G_D77_3_2_1_3
  system shall log performance_parameters of vehicle
  stakeholders vehicle, om_system, connected_vehicle_system
  linked_to "G.D77.3.2.1.3"

// Row 089 | source_id: G.D77.3.2.2.10 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_089_ID_G_D77_3_2_2_10
  system shall warn driver
  stakeholders driver, user, connected_vehicle_system, asd
  linked_to "G.D77.3.2.2.10"

// Row 090 | source_id: G.D77.3.2.3.9 | definability: partially
// Issue: Quantitative/timing/capacity constraints are not first-class in Value-DSL and were
// preserved only implicitly.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_090_ID_G_D77_3_2_3_9
  system shall monitor location_data of vehicle
  stakeholders vehicle, location, connected_vehicle_system, asd, host_vehicle
  linked_to "G.D77.3.2.3.9"

// Row 091 | source_id: G.D77.3.2.5.12 | definability: partially
// Issue: Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_091_ID_G_D77_3_2_5_12
  system shall warn driver
  stakeholders driver, vehicle, connected_vehicle_system, nyc_cvpd_subsystem
  linked_to "G.D77.3.2.5.12"

// Row 092 | source_id: G.D77.3.8.3.7.1 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_092_ID_G_D77_3_8_3_7_1
  if pedestrian pedestrian_condition_detected
  system shall warn transit_vehicle_operator using display
  stakeholders pedestrian, vehicle, connected_vehicle_system, transit_vehicle_operator
  linked_to "G.D77.3.8.3.7.1"

// Row 093 | source_id: G.D77.3.8.3.7.2 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_093_ID_G_D77_3_8_3_7_2
  where condition_from_source_93
  system shall inform connected_vehicle_system
  stakeholders pedestrian, connected_vehicle_system
  linked_to "G.D77.3.8.3.7.2"

// Row 094 | source_id: G.D77.3.8.3.7.4 | definability: partially
// Issue: Quantitative/timing/capacity constraints are not first-class in Value-DSL and were
// preserved only implicitly.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_094_ID_G_D77_3_8_3_7_4
  system shall warn connected_vehicle_system
  stakeholders pedestrian, om_system, connected_vehicle_system
  linked_to "G.D77.3.8.3.7.4"

// Row 095 | source_id: G.D77.3.8.4.1.1 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_095_ID_G_D77_3_8_4_1_1
  system shall detect pedestrian_presence of pedestrian
  stakeholders pedestrian, connected_vehicle_system
  linked_to "G.D77.3.8.4.1.1"

// Row 096 | source_id: G.D77.3.8.4.1.5 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_096_ID_G_D77_3_8_4_1_5
  system shall inform connected_vehicle_system
  stakeholders pedestrian, om_system, connected_vehicle_system
  linked_to "G.D77.3.8.4.1.5"

// Row 097 | source_id: G.D77.3.8.4.1.7 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_097_ID_G_D77_3_8_4_1_7
  system shall inform connected_vehicle_system
  stakeholders pedestrian, om_system, trip, connected_vehicle_system
  linked_to "G.D77.3.8.4.1.7"

// Row 098 | source_id: G.D77.3.8.4.1.19 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_098_ID_G_D77_3_8_4_1_19
  system shall inform connected_vehicle_system
  stakeholders pedestrian, connected_vehicle_system
  linked_to "G.D77.3.8.4.1.19"

// Row 099 | source_id: G.D81.3.1 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_099_ID_G_D81_3_1
  system shall measure blood_sugar_level of om_system
  stakeholders om_system
  linked_to "G.D81.3.1"

// Row 100 | source_id: G.D81.3.2 | definability: partially
// Issue: Quantitative/timing/capacity constraints are not first-class in Value-DSL and were
// preserved only implicitly.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_100_ID_G_D81_3_2
  where condition_from_source_100
  system shall measure blood_sugar_level of application_system
  stakeholders application_system
  linked_to "G.D81.3.2"

// Row 101 | source_id: G.D82.6.1.4 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_101_ID_G_D82_6_1_4
  system shall detect location_data of location
  stakeholders location, patient, ehr_system
  linked_to "G.D82.6.1.4"

// Row 102 | source_id: G.D82.6.1.12 | definability: no
// Issue: Generic voice-recognition capability has no stakeholder/requirement action in Value-DSL
// without inventing an unsupported capability construct.
// Note: Kept as comment only in the DSL file.
// Unmapped source requirement: The system has capability to have voice recognition.

// Row 103 | source_id: G.D82.6.1.16 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_103_ID_G_D82_6_1_16
  system shall inform user
  stakeholders ehr_system, user
  linked_to "G.D82.6.1.16"

// Row 104 | source_id: G.D82.6.1.19 | definability: partially
// Issue: Generic capability/support wording was approximated as a record requirement; Value-DSL
// lacks native capability/provide/support constructs.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_104_ID_G_D82_6_1_19
  system shall record demographic_data of om_system
  stakeholders om_system, person, ehr_system
  linked_to "G.D82.6.1.19"

// Row 105 | source_id: G.D82.6.1.21 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb. Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_105_ID_G_D82_6_1_21
  system shall track manage_track_report_user of patient
  stakeholders patient, ehr_system, user
  linked_to "G.D82.6.1.21"

// Row 106 | source_id: G.D82.6.1.41 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb. Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_106_ID_G_D82_6_1_41
  system shall track patient_id_changes of patient
  stakeholders patient, ehr_system
  linked_to "G.D82.6.1.41"

// Row 107 | source_id: G.D82.6.2.87 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_107_ID_G_D82_6_2_87
  system shall inform ehr_system using display
  stakeholders om_system, location, patient, ehr_system
  linked_to "G.D82.6.2.87"

// Row 108 | source_id: G.D82.6.4.135 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_108_ID_G_D82_6_4_135
  system shall inform user
  stakeholders vehicle, ehr_system, user
  linked_to "G.D82.6.4.135"

// Row 109 | source_id: G.D82.6.4.159 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb. Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_109_ID_G_D82_6_4_159
  system shall track patient_problems of patient
  stakeholders patient, ehr_system
  linked_to "G.D82.6.4.159"

// Row 110 | source_id: G.D82.6.4.167 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_110_ID_G_D82_6_4_167
  system shall inform nurse
  stakeholders ehr_system, nurse
  linked_to "G.D82.6.4.167"

// Row 111 | source_id: G.D82.6.5.224 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action. Compound requirement was collapsed to one primary Value-DSL
// action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_111_ID_G_D82_6_5_224
  where condition_from_source_111
  system shall inform caregiver
  stakeholders vehicle, om_system, ehr_system, caregiver
  linked_to "G.D82.6.5.224"

// Row 112 | source_id: G.D82.6.6.278 | definability: partially
// Issue: Unsupported analysis/capability verb was approximated with the nearest supported
// monitoring verb. Compound requirement was collapsed to one primary Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_112_ID_G_D82_6_6_278
  system shall track consultation_counts of ehr_system
  stakeholders ehr_system, physician
  linked_to "G.D82.6.6.278"

// Row 113 | source_id: G.D82.6.6.281 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited. Compound requirement was collapsed to one primary
// Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_113_ID_G_D82_6_6_281
  system shall inform physician
  stakeholders ehr_system, physician
  linked_to "G.D82.6.6.281"

// Row 114 | source_id: G.D82.6.6.284 | definability: partially
// Issue: Display/provide/present/list capability was approximated as an inform action because
// Value-DSL notification verbs are limited.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_114_ID_G_D82_6_6_284
  system shall inform ehr_system
  stakeholders patient, ehr_system
  linked_to "G.D82.6.6.284"

// Row 115 | source_id: G.D82.6.9.427 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_115_ID_G_D82_6_9_427
  system shall track turnaround_times of ehr_system
  stakeholders ehr_system, physician
  linked_to "G.D82.6.9.427"

// Row 116 | source_id: G.D82.6.12.504 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_116_ID_G_D82_6_12_504
  system shall log physiologic_monitor_data of component
  stakeholders component, om_system, ehr_system, icu_component, physiologic_monitor
  linked_to "G.D82.6.12.504"

// Row 117 | source_id: G.D82.6.12.505 | definability: yes
// Note: Direct fit to the supported stakeholder/requirement constructs.
requirement REQ_117_ID_G_D82_6_12_505
  system shall capture bedside_monitor_data of component
  stakeholders component, om_system, ehr_system, icu_component, bedside_monitor
  linked_to "G.D82.6.12.505"

// Row 118 | source_id: G.D82.6.12.511 | definability: partially
// Issue: Some source semantics use unsupported capability/display/calculation wording; mapped to
// the nearest Value-DSL action.
// Note: Core stakeholder/action intent represented; secondary details may require separate
// requirements or unsupported DSL extensions.
requirement REQ_118_ID_G_D82_6_12_511
  system shall inform ehr_system
  stakeholders patient, ehr_system
  linked_to "G.D82.6.12.511"
