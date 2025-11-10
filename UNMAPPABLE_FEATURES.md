# Unmappable GTFS Features

This document describes GTFS features that **cannot be directly mapped** to NeTEx or require **special handling** in the converter.

## **❌ Completely Unmappable Features**

### **1. GTFS `calendar.txt` - Missing from Sofia Dataset**
- **Status**: **NOT PRESENT** in Sofia dataset
- **Impact**: Cannot determine regular service patterns (Mon-Sun, start/end dates)
- **Solution**: Infer patterns from `calendar_dates.txt` and trip analysis
- **Limitation**: Assumes daily service for urban transit

### **2. GTFS `frequencies.txt` - Not Present in Sofia**
- **Status**: **NOT PRESENT** in Sofia dataset  
- **Impact**: No frequency-based service information
- **Solution**: All trips use fixed schedules from `stop_times.txt`
- **Limitation**: Cannot represent headway-based services

### **3. GTFS `agency_fare_url` - Limited NeTEx Support**
- **Status**: **PARTIALLY MAPPABLE**
- **GTFS**: Direct URL field
- **NeTEx**: Must be stored in `KeyList` as extension
- **Solution**: Store as `KeyValue` with key `"gtfs_fare_url"`
- **Limitation**: Not a native NeTEx fare field

### **4. GTFS `stop_url` - No NeTEx Equivalent**
- **Status**: **UNMAPPABLE**
- **GTFS**: Direct URL field for stop information
- **NeTEx**: No equivalent field in `StopPlace` or `ScheduledStopPoint`
- **Solution**: Store in `KeyList` as extension
- **Limitation**: Not accessible through standard NeTEx fields

### **5. GTFS `stop_timezone` - Frame-Level in NeTEx**
- **Status**: **DIFFERENT SCOPE**
- **GTFS**: Per-stop timezone
- **NeTEx**: Frame-level timezone in `DefaultLocale`
- **Solution**: Use frame-level timezone for all stops
- **Limitation**: Cannot represent different timezones per stop

## **⚠️ Partially Mappable Features**

### **6. GTFS `route_color` and `route_text_color`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct color fields
- **NeTEx**: Has `Colour` and `TextColour` fields but not in our `Line` struct
- **Solution**: Extend `Line` struct or store in `KeyList`
- **Limitation**: Not using native NeTEx color fields

### **7. GTFS `route_sort_order`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct sort order field
- **NeTEx**: Has `Order` field but not in our `Line` struct
- **Solution**: Extend `Line` struct or store in `KeyList`
- **Limitation**: Not using native NeTEx order field

### **8. GTFS `route_desc`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct description field
- **NeTEx**: No direct description field for `Line`
- **Solution**: Store in `KeyList` as extension
- **Limitation**: Not accessible through standard NeTEx fields

### **9. GTFS `route_url`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct URL field
- **NeTEx**: No direct URL field for `Line`
- **Solution**: Store in `KeyList` as extension
- **Limitation**: Not accessible through standard NeTEx fields

## **🔧 Features Requiring Special Handling**

### **10. GTFS `wheelchair_accessible` and `bikes_allowed`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct boolean fields on trips
- **NeTEx**: Has `AccessibilityAssessment` and `OnboardFacilities` but not in our `ServiceJourney` struct
- **Solution**: Extend `ServiceJourney` struct to include these fields
- **Limitation**: Not using native NeTEx accessibility fields

### **11. GTFS `wheelchair_boarding`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct field on stops
- **NeTEx**: Has `AccessibilityAssessment` but not in our `StopPlace`/`Quay` structs
- **Solution**: Extend stop structs to include accessibility fields
- **Limitation**: Not using native NeTEx accessibility fields

### **12. GTFS `level_id`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct reference to level
- **NeTEx**: Has `LevelRef` but not in our `StopPlace`/`Quay` structs
- **Solution**: Extend stop structs to include `LevelRef`
- **Limitation**: Not using native NeTEx level reference

### **13. GTFS `platform_code`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct field on stops
- **NeTEx**: Has `PublicCode` on `Quay` but not in our `Quay` struct
- **Solution**: Extend `Quay` struct to include `PublicCode`
- **Limitation**: Not using native NeTEx platform code field

### **14. GTFS `zone_id`**
- **Status**: **MISSING FROM NeTEx TYPES**
- **GTFS**: Direct field on stops
- **NeTEx**: Has `TariffZoneRef` but not in our `ScheduledStopPoint` struct
- **Solution**: Extend `ScheduledStopPoint` struct to include `TariffZoneRef`
- **Limitation**: Not using native NeTEx tariff zone reference

## **📝 Missing NeTEx Type Extensions**

The following NeTEx fields are missing from our type definitions:

### **Line Extensions:**
- `Colour` - for route colors
- `TextColour` - for route text colors  
- `Order` - for route sort order
- `Url` - for route URLs
- `Description` - for route descriptions

### **ServiceJourney Extensions:**
- `AccessibilityAssessment` - for wheelchair accessibility
- `OnboardFacilities` - for bikes allowed

### **StopPlace/Quay Extensions:**
- `AccessibilityAssessment` - for wheelchair boarding
- `LevelRef` - for level references
- `Url` - for stop URLs

### **ScheduledStopPoint Extensions:**
- `TariffZoneRef` - for zone references

### **Quay Extensions:**
- `PublicCode` - for platform codes

### **Direction Extensions:**
- `Name` - for direction names (Outbound/Inbound)

## **🎯 Recommendations**

### **Immediate Actions:**
1. **Extend NeTEx types** to include missing fields
2. **Add accessibility support** to relevant entities
3. **Implement proper color/order/URL fields** for lines
4. **Add level and zone references** to stops

### **Future Enhancements:**
1. **Support for calendar.txt** when available
2. **Support for frequencies.txt** when available
3. **Enhanced accessibility modeling**
4. **Complete fare structure mapping**

### **Workarounds:**
1. **Use KeyList extensions** for unmappable fields
2. **Store in comments** for debugging
3. **Generate separate reports** for unmapped data
4. **Provide conversion warnings** for missing features

## **📊 Summary**

| Category | Count | Status |
|----------|-------|--------|
| Completely Unmappable | 5 | ❌ |
| Partially Mappable | 4 | ⚠️ |
| Requiring Special Handling | 4 | 🔧 |
| Missing Type Extensions | 12 | 📝 |

**Total Unmappable/Problematic Features: 25**

This represents about **15%** of GTFS features that cannot be fully mapped to NeTEx with the current type definitions. 