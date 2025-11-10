# GTFS to NeTEx Mapping Reference  
**Based on DATA4PT GTFS-NeTEx Mapping (Version June 2024)**

---

## **Overview**
This document provides a comprehensive mapping reference between **GTFS (General Transit Feed Specification)** and **NeTEx (Network Timetable Exchange)** based on the official **DATA4PT** specification.

---

## **Key Principles**
- **Conceptual Separation**: NeTEx separates logical (scheduled) and physical (infrastructure) concepts  
- **One-to-One Mapping**: Each GTFS attribute maps to a specific NeTEx attribute  
- **Split Records**: One GTFS record may map to multiple NeTEx entities  
- **Explicit Objects**: NeTEx uses explicit objects without overloading semantics  

---

## **Core Entity Mappings**

### **1. GTFS agency → NeTEx OPERATOR**

| GTFS Attribute    | NeTEx Element | NeTEx Attribute          | Type             | Notes         |
|-------------------|--------------|--------------------------|-----------------|---------------|
| agency_id         | Operator     | id                      | OperatorIdType |               |
| agency_name       | Operator     | Name                    | MultilingualString |          |
| agency_url        | Operator     | ContactDetails.Url      | xsd:anyURI      |               |
| agency_timezone   | Operator     | Timezone                | xsd:string      |               |
| agency_lang       | Operator     | DefaultLanguage         | xsd:lang        |               |
| agency_phone      | Operator     | ContactDetails.Phone    | PhoneNumber     |               |
| agency_email      | Operator     | ContactDetails.Email    | Email           |               |
| agency_fare_url   | Operator     | Keylist.gtfs_fare_url   | xsd:string      | Extension     |

---

### **2. GTFS stops → NeTEx SCHEDULED STOP POINT / STOP PLACE**

**GTFS overloads the stops record for multiple concepts. NeTEx separates these:**

| GTFS location_type | NeTEx Concept                     | Description                                 |
|---------------------|-----------------------------------|---------------------------------------------|
| 0 (Stop/Platform)  | SCHEDULED STOP POINT + STOP PLACE + QUAY | Timetable point + physical infrastructure |
| 1 (Station)        | STOP PLACE                       | Physical station/terminal                  |
| 2 (Entrance/Exit)  | ENTRANCE                         | Station entrance                            |
| 3 (Generic Node)   | ACCESS SPACE                     | Navigation node                             |
| 4 (Boarding Area)  | BOARDING POSITION                | Specific boarding location                 |

#### Stop Attributes Mapping

| GTFS Attribute   | NeTEx Element                    | NeTEx Attribute        | Notes            |
|------------------|---------------------------------|------------------------|------------------|
| stop_id          | SCHEDULED STOP POINT / STOP PLACE | id                   | Context dependent |
| stop_code        | SCHEDULED STOP POINT            | PublicCode            |                  |
| stop_name        | SCHEDULED STOP POINT / STOP PLACE | Name                 |                  |
| stop_desc        | SCHEDULED STOP POINT / STOP PLACE | Description          |                  |
| stop_lat         | STOP PLACE / QUAY               | Centroid.Location.Latitude |           |
| stop_lon         | STOP PLACE / QUAY               | Centroid.Location.Longitude |          |
| zone_id          | SCHEDULED STOP POINT            | TariffZoneRef         |                  |
| stop_url         | STOP PLACE                      | Url                   |                  |
| location_type    | -                               | Determines entity type |                  |
| parent_station   | QUAY / ENTRANCE                 | ParentSiteRef         |                  |
| stop_timezone    | STOP PLACE                      | Timezone              |                  |
| wheelchair_boarding | STOP PLACE / QUAY            | AccessibilityAssessment |                 |
| level_id         | STOP PLACE / QUAY               | LevelRef              |                  |
| platform_code    | QUAY                            | PublicCode            |                  |

---

### **3. GTFS routes → NeTEx LINE**

| GTFS Attribute    | NeTEx Element | NeTEx Attribute       | Notes                      |
|-------------------|--------------|-----------------------|---------------------------|
| route_id          | LINE         | id                   |                           |
| agency_id         | LINE         | OperatorRef          | Reference to OPERATOR    |
| route_short_name  | LINE         | ShortName / PublicCode |                        |
| route_long_name   | LINE         | Name                 |                           |
| route_desc        | LINE         | Description          |                           |
| route_type        | LINE         | TransportMode        | See mode mapping below    |
| route_url         | LINE         | Url                  |                           |
| route_color       | LINE         | Colour               |                           |
| route_text_color  | LINE         | TextColour           |                           |
| route_sort_order  | LINE         | Order                |                           |

#### Route Type to Transport Mode Mapping

| GTFS route_type | NeTEx TransportMode |
|-----------------|----------------------|
| 0              | tram                |
| 1              | metro               |
| 2              | rail                |
| 3              | bus                 |
| 4              | water               |
| 5              | cableway            |
| 6              | cableway            |
| 7              | funicular           |
| 11             | trolleyBus          |
| 12             | rail                |

---

### **4. GTFS trips → NeTEx SERVICE JOURNEY**

| GTFS Attribute      | NeTEx Element     | NeTEx Attribute            | Notes                           |
|----------------------|-------------------|---------------------------|--------------------------------|
| route_id             | SERVICE JOURNEY  | LineRef                  |                                |
| service_id           | SERVICE JOURNEY  | DayTypeRef               | Via calendar mapping           |
| trip_id              | SERVICE JOURNEY  | id                       |                                |
| trip_headsign        | SERVICE JOURNEY  | DestinationDisplayRef     |                                |
| trip_short_name      | SERVICE JOURNEY  | PublicCode               |                                |
| direction_id         | SERVICE JOURNEY  | DirectionRef             |                                |
| block_id             | SERVICE JOURNEY  | BlockRef                 |                                |
| shape_id             | SERVICE JOURNEY  | JourneyPatternRef         | Via shape mapping             |
| wheelchair_accessible| SERVICE JOURNEY  | AccessibilityAssessment   |                                |
| bikes_allowed        | SERVICE JOURNEY  | OnboardFacilities         |                                |

---

### **5. GTFS stop_times → NeTEx CALL**

GTFS stop_times become **TIMETABLED PASSING TIME** elements within SERVICE JOURNEY:

| GTFS Attribute   | NeTEx Element             | NeTEx Attribute         | Notes               |
|-------------------|--------------------------|--------------------------|----------------------|
| trip_id           | -                       | Parent SERVICE JOURNEY   |                     |
| arrival_time      | TIMETABLED PASSING TIME | ArrivalTime             |                     |
| departure_time    | TIMETABLED PASSING TIME | DepartureTime           |                     |
| stop_id           | TIMETABLED PASSING TIME | ScheduledStopPointRef   |                     |
| stop_sequence     | TIMETABLED PASSING TIME | Order                   |                     |
| stop_headsign     | CALL                    | DestinationDisplayRef    |                     |
| pickup_type       | CALL                    | RequestStop (ForBoarding)|                     |
| drop_off_type     | CALL                    | RequestStop (ForAlighting)|                   |
| shape_dist_traveled | POINT IN JOURNEY PATTERN | DistanceFromStart    |                     |
| timepoint         | TIMETABLED PASSING TIME | IsFlexible              | Inverted logic      |

---

### **6. GTFS calendar + calendar_dates → NeTEx DAY TYPE + DAY TYPE ASSIGNMENT**

#### Calendar Mapping

| GTFS Attribute    | NeTEx Element       | NeTEx Attribute              | Notes                     |
|--------------------|---------------------|-----------------------------|---------------------------|
| service_id         | DAY TYPE           | id                          |                           |
| monday-sunday      | DAY TYPE           | Properties.PropertyOfDay    |                           |
| start_date         | DAY TYPE ASSIGNMENT| OperatingPeriod.FromDate    |                           |
| end_date           | DAY TYPE ASSIGNMENT| OperatingPeriod.ToDate      |                           |

#### Calendar Dates Mapping

| GTFS Attribute   | NeTEx Element       | NeTEx Attribute             | Notes                         |
|-------------------|---------------------|----------------------------|-------------------------------|
| service_id        | DAY TYPE ASSIGNMENT| DayTypeRef                |                               |
| date              | DAY TYPE ASSIGNMENT| Date or OperatingDay      |                               |
| exception_type    | DAY TYPE ASSIGNMENT| IsAvailable               | 1=true, 2=false              |

---

### **7. GTFS transfers → NeTEx CONNECTION + INTERCHANGE RULE**

| GTFS Attribute    | NeTEx Element    | NeTEx Attribute           | Notes                      |
|--------------------|-----------------|---------------------------|---------------------------|
| from_stop_id       | CONNECTION      | From.ScheduledStopPointRef|                           |
| to_stop_id         | CONNECTION      | To.ScheduledStopPointRef  |                           |
| transfer_type      | INTERCHANGE RULE| RestrictionType           | See mapping below          |
| min_transfer_time  | CONNECTION      | TransferDuration.DefaultDuration |                     |

#### Transfer Type Mapping

| GTFS transfer_type | NeTEx RestrictionType     |
|---------------------|---------------------------|
| 0                  | recommendedTransfer       |
| 1                  | noTransfer               |
| 2                  | guaranteedTransfer       |
| 3                  | cannotTransfer           |

---

### **8. Additional Files**

#### GTFS shapes → NeTEx ROUTE LINK + POINT ON LINK

| GTFS Attribute       | NeTEx Element      | NeTEx Attribute          |
|-----------------------|-------------------|--------------------------|
| shape_id             | ROUTE            | id                      |
| shape_pt_lat         | POINT ON LINK    | Location.Latitude       |
| shape_pt_lon         | POINT ON LINK    | Location.Longitude      |
| shape_pt_sequence    | POINT ON LINK    | Order                   |
| shape_dist_traveled  | POINT ON LINK    | DistanceFromStart       |

#### GTFS levels → NeTEx LEVEL

| GTFS Attribute    | NeTEx Element | NeTEx Attribute  |
|-------------------|--------------|------------------|
| level_id          | LEVEL        | id              |
| level_index       | LEVEL        | PublicCode      |
| level_name        | LEVEL        | Name            |

#### GTFS pathways → NeTEx PATH LINK

| GTFS Attribute         | NeTEx Element    | NeTEx Attribute                 |
|-------------------------|-----------------|---------------------------------|
| pathway_id             | PATH LINK      | id                              |
| from_stop_id           | PATH LINK      | From.PlaceRef                   |
| to_stop_id             | PATH LINK      | To.PlaceRef                     |
| pathway_mode           | PATH LINK      | TransitionType                  |
| is_bidirectional       | PATH LINK      | BothWays                        |
| length                 | PATH LINK      | Distance                         |
| traversal_time         | PATH LINK      | TransferDuration.DefaultDuration|
| stair_count            | PATH LINK      | NumberOfSteps                   |
| max_slope              | PATH LINK      | MaximumGradient                 |
| min_width              | PATH LINK      | MinimumWidth                    |
| signposted_as          | SIGN EQUIPMENT | SignContent                     |
| reversed_signposted_as | SIGN EQUIPMENT | ReverseSignContent              |

#### GTFS feed_info → NeTEx DATA SOURCE

| GTFS Attribute        | NeTEx Element     | NeTEx Attribute                      |
|------------------------|------------------|--------------------------------------|
| feed_publisher_name   | DATA SOURCE     | Name                                 |
| feed_publisher_url    | DATA SOURCE     | Url                                  |
| feed_lang             | DATA SOURCE     | DefaultLanguage                      |
| feed_start_date       | COMPOSITE FRAME | ValidityCondition.FromDate          |
| feed_end_date         | COMPOSITE FRAME | ValidityCondition.ToDate            |
| feed_version          | COMPOSITE FRAME | Version                              |
| feed_contact_email    | DATA SOURCE     | ContactDetails.Email                |
| feed_contact_url      | DATA SOURCE     | ContactDetails.Url                  |

---

## **ID Generation Guidelines**
- **General Pattern**:  
  Use meaningful prefixes:  
  `operator_`, `line_`, `stop_place_`, `service_journey_`

- **Preserve original GTFS IDs where possible**
- **Ensure uniqueness within scope**

**Examples:**
- GTFS `agency_id`: `A` → NeTEx: `operator_A`
- GTFS `stop_id`: `12345` → NeTEx: `stop_place_12345` or `scheduled_stop_point_12345`
- GTFS `route_id`: `R1` → NeTEx: `line_R1`
- GTFS `trip_id`: `T123` → NeTEx: `service_journey_T123`

---

## **Frame Organization**
NeTEx organizes data into **frames**:
- **GeneralFrame**: Infrastructure (STOP PLACEs, LEVELs)
- **ResourceFrame**: Organizations (OPERATORs, AUTHORITYs)
- **ServiceFrame**: Network topology (LINEs, ROUTEs, JOURNEY PATTERNs)
- **TimetableFrame**: Schedules (SERVICE JOURNEYs, PASSING TIMEs)
- **ServiceCalendarFrame**: Operating days (DAY TYPEs, OPERATING PERIODs)

---

## **Key Differences**
- **Separation of Concerns**: NeTEx separates scheduled vs physical stops
- **Explicit Relationships**: NeTEx uses explicit references instead of implicit relationships
- **Versioning**: NeTEx supports fine-grained versioning at element level
- **Multilingual Support**: Built-in multilingual text support
- **Accessibility**: Comprehensive accessibility modeling

---

## **Implementation Notes**
- Always create both **SCHEDULED STOP POINT** and **STOP PLACE** for GTFS stops  
- Use **STOP ASSIGNMENT** to link scheduled and physical stops  
- Generate **JOURNEY PATTERN** from unique stop sequences  
- Create **ROUTE** objects to represent the geographic path  
- Implement proper referential integrity with `ref` attributes  
- Use appropriate NeTEx frames for organization  
- Apply consistent ID generation patterns  

---

# GTFS to NeTEx Mapping Reference - CORRECTED
**Based on DATA4PT GTFS-NeTEx Mapping (Version June 2024)**

---

## **Missing Mappings Added:**

### **9. GTFS fare_attributes → NeTEx FARE FRAME**

| GTFS Attribute        | NeTEx Element     | NeTEx Attribute                      |
|------------------------|------------------|--------------------------------------|
| fare_id               | TARIFF           | id                                   |
| price                 | TARIFF           | Amount                               |
| currency_type         | TARIFF           | Currency                             |
| payment_method        | TARIFF           | PaymentMethod                        |
| transfers             | TARIFF           | MaximumNumberOfTransfers             |
| agency_id             | TARIFF           | OperatorRef                          |
| transfer_duration     | TARIFF           | TransferDuration                     |

### **10. GTFS translations → NeTEx MULTILINGUAL SUPPORT**

| GTFS Attribute        | NeTEx Element     | NeTEx Attribute                      |
|------------------------|------------------|--------------------------------------|
| table_name            | -                | Determines target entity             |
| field_name            | -                | Determines target attribute          |
| language              | -                | Language code                        |
| translation           | Various          | MultilingualString.Value             |
| record_id             | -                | Target entity ID                     |
| record_sub_id         | -                | Target sub-entity ID                 |
| field_value           | -                | Target field value                   |

### **11. GTFS direction_id → NeTEx DIRECTION**

| GTFS Attribute        | NeTEx Element     | NeTEx Attribute                      |
|------------------------|------------------|--------------------------------------|
| direction_id (0)      | DIRECTION        | id="direction_0"                     |
| direction_id (1)      | DIRECTION        | id="direction_1"                     |
| -                     | DIRECTION        | Name="Outbound" / "Inbound"          |

### **12. GTFS shapes → NeTEx ROUTE LINK (CORRECTED)**

| GTFS Attribute        | NeTEx Element     | NeTEx Attribute                      |
|------------------------|------------------|--------------------------------------|
| shape_id              | ROUTE            | id                                   |
| shape_pt_lat          | LINE STRING      | pos (lat,lon format)                 |
| shape_pt_lon          | LINE STRING      | pos (lat,lon format)                 |
| shape_pt_sequence     | POINT ON LINK    | Order                                |
| shape_dist_traveled   | POINT ON LINK    | DistanceFromStart                    |

**RouteLink Structure:**
```xml
<RouteLink id="route_link_1">
  <LineString>
    <pos>42.123 23.456</pos>
    <pos>42.124 23.457</pos>
  </LineString>
  <FromPointRef ref="route_point_1"/>
  <ToPointRef ref="route_point_2"/>
</RouteLink>
```

### **13. JourneyPattern Generation**

**From GTFS trip + stop_times → NeTEx ServiceJourneyPattern:**

| GTFS Source           | NeTEx Element     | Generation Rule                      |
|------------------------|------------------|--------------------------------------|
| trip_id               | SERVICE JOURNEY PATTERN | id="journey_pattern_{trip_id}"    |
| stop_times sequence   | STOP POINT IN JOURNEY PATTERN | Order from stop_sequence        |
| stop_id               | SCHEDULED STOP POINT REF | Reference to stop point         |
| pickup_type           | FOR BOARDING     | Boolean mapping                      |
| drop_off_type         | FOR ALIGHTING    | Boolean mapping                      |

### **14. VehicleType Generation**

**Since GTFS lacks vehicle types, generate from route_type:**

| GTFS route_type       | NeTEx VehicleType | Name                                |
|------------------------|------------------|--------------------------------------|
| 0 (Tram)              | VEHICLE TYPE     | "Tram"                              |
| 1 (Metro)             | VEHICLE TYPE     | "Metro"                             |
| 2 (Rail)              | VEHICLE TYPE     | "Train"                             |
| 3 (Bus)               | VEHICLE TYPE     | "Bus"                               |
| 11 (Trolleybus)       | VEHICLE TYPE     | "Trolleybus"                        |

### **15. Missing Stop Assignment Mapping**

| GTFS Relationship      | NeTEx Element     | NeTEx Attribute                      |
|------------------------|------------------|--------------------------------------|
| parent_station         | PASSENGER STOP ASSIGNMENT | StopPlaceRef                    |
| stop_id                | PASSENGER STOP ASSIGNMENT | ScheduledStopPointRef          |
| platform_code          | PASSENGER STOP ASSIGNMENT | QuayRef                         |

---

## **Implementation Corrections:**

### **Calendar Handling (Missing calendar.txt)**
```go
// Infer calendar from calendar_dates.txt and trips.txt
func inferCalendar(calendarDates []CalendarDate, trips []Trip) []DayType {
    // Create default daily service pattern
    // Use calendar_dates.txt for exceptions
    // Generate DayType and DayTypeAssignment
}
```

### **Complete Frame Structure**
```xml
<PublicationDelivery>
  <dataObjects>
    <CompositeFrame>
      <frames>
        <ResourceFrame>        <!-- Operators, Authorities -->
        <ServiceCalendarFrame> <!-- DayTypes, OperatingPeriods -->
        <ServiceFrame>         <!-- Lines, Routes, JourneyPatterns -->
        <SiteFrame>           <!-- StopPlaces, Quays -->
        <TimetableFrame>      <!-- ServiceJourneys, PassingTimes -->
      </frames>
    </CompositeFrame>
  </dataObjects>
</PublicationDelivery>
```

---

## **Converter Implementation Steps:**

1. **Load all GTFS files**
2. **Generate missing entities** (Directions, VehicleTypes, JourneyPatterns)
3. **Transform stops** (split into ScheduledStopPoint + StopPlace + Quay)
4. **Create StopAssignments** (link scheduled and physical stops)
5. **Generate RouteLinks** (from shapes)
6. **Create JourneyPatterns** (from trip stop sequences)
7. **Map all entities** to NeTEx structure
8. **Organize into frames**
9. **Generate XML output**

---