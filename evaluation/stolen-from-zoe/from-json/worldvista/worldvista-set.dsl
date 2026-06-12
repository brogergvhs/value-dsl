stakeholder Patients
stakeholder Product_Owners
stakeholder Relatives
stakeholder Doctors
stakeholder System
stakeholder User

// Source wv053:
// The System shall determine 'if a user is authorized to sign and order'.
requirement wv053
system shall detect 'authorization to sign and order' of User
stakeholders Patients, Product_Owners, Doctors, Relatives, User
linked_to wv053_source

// Source wv101:
// The System shall log 'access' to 'sensitive records' with the aim of 'enabling a site to follow up on a record accessed by a remote location'.
requirement wv101
system shall log 'access to sensitive records' of Patients using 'follow up on remote record access'
stakeholders Patients, Product_Owners
linked_to wv101_source

// Source wv114:
// The System shall display 'patient demographics, allergies, adverse reactions, postings, clinical reminders, current medications, problems and appointments' of a Patient to a Patient and Doctor.
requirement wv114
system shall notify Doctors using 'patient demographics, allergies, adverse reactions, postings, clinical reminders, current medications, problems and appointments'
stakeholders Patients, Product_Owners, Relatives, Doctors
linked_to wv114_source

// Source wv115:
// The System shall assess 'orders' for 'duplicates, drug-drug interactions, and other user criteria'.
requirement wv115
system shall detect 'duplicates, drug-drug interactions, and other user criteria' of Doctors using orders
stakeholders Patients, Relatives, Product_Owners, Doctors
linked_to wv115_source

// Source wv123a:
// The System shall collect 'encounter data for workload credit' by the means of 'Ambulatory Care Data Capture project'.
requirement wv123a
system shall collect 'encounter data for workload credit' of Patients using 'Ambulatory Care Data Capture project'
stakeholders Patients, Relatives, Product_Owners
linked_to wv123a_source

// Source wv123b:
// The System shall collect 'clinically relevant data' with the aim of 'creating reminders and reports' by the means of 'Ambulatory Care Data Capture project'.
requirement wv123b
system shall collect 'clinically relevant data' of Patients using 'Ambulatory Care Data Capture project for reminders and reports'
stakeholders Patients, Relatives, Product_Owners, Doctors
linked_to wv123b_source

// Source wv126a:
// The System shall produce 'Health Summary, Vitals Cumulative, Nutritional Assessment, Daily Order Summary and Order Summary reports' from 'a Patient's data'.
requirement wv126a
system shall record 'Health Summary, Vitals Cumulative, Nutritional Assessment, Daily Order Summary and Order Summary reports' of Patients using 'patient data'
stakeholders Doctors, Relatives, Patients
linked_to wv126a_source

// Source wv126b:
// The System shall allow 'customization of Health Summary, Vitals Cumulative, Nutritional Assessment, Daily Order Summary and Order Summary reports'.
requirement wv126b
system shall notify Doctors using 'customization of Health Summary, Vitals Cumulative, Nutritional Assessment, Daily Order Summary and Order Summary reports'
stakeholders Doctors, Relatives, Patients
linked_to wv126b_source

// Source wv128:
// The System shall store 'risk, social and medical factors including tobacco use, alcohol use, drug use, occupational environment, marital status, occupation, religious preference, ethnicity, healthcare surrogate and guardian'.
requirement wv128
system shall store 'risk, social and medical factors including tobacco use, alcohol use, drug use, occupational environment, marital status, occupation, religious preference, ethnicity, healthcare surrogate and guardian' of Patients
stakeholders Patients, Relatives, Product_Owners, Doctors
linked_to wv128_source

// Source wv137:
// The System shall include 'non-Standard Laboratory and Medication orderable files'.
requirement wv137
system shall record 'non-Standard Laboratory and Medication orderable files' of System
stakeholders Patients, Relatives, Doctors, System
linked_to wv137_source
