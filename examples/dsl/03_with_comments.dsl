// Stakeholders
stakeholder Worker
stakeholder Manager

// Value declarations
value privacy_pref = 1.58, 0.91
value authority_pref = 0.02, 0.88

// Requirement block
requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Manager
