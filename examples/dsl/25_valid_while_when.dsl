stakeholder Worker
stakeholder Supervisor

requirement R1
while Worker is in RestrictedZone
when Worker enters RestrictedZone
system shall track location of Worker using Camera
stakeholders Worker, Supervisor
priority high
retention short_term
linked_to Hazard_RestrictedZone

assignment R1
Worker -> privacy
Supervisor -> safety
