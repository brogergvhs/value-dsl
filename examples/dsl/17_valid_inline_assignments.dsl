stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91
value authority_pref = 0.02, 0.88

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Manager

requirement R2
when Manager reports Incident
system shall notify Worker using AudioAlarm
stakeholders Worker, Manager

assignment R1
Worker -> privacy_pref
Manager -> authority

assignment R2
Worker -> privacy_pref
Manager -> authority_pref
