## [1.4.1] (2024-05-06)

* The field used to sample glucose has been changed.

```
was: dateString - string
stato: date - unixMilliseconds
```

## [1.2.1-beta] (2023-10-17)

* Added experimental commands for creating and viewing some Nightscout objects (hidden commands).
* Maken more informative output of some errors.

## [1.2.0-beta] (2023-09-30)

### New

* Added the ability to export long-acting insulin Lantus and Toujeo from xDrip+ (xDrip+ must be configured for multiple insulin profiles. E.g. very-fast Fiasp insulin and long-acting Lantus insulin)

## [1.1.0-beta] (2023-09-26)

### Fix

* corrected type for carbohydrates (int -> float64)

### New

* Added commands for generating, editing and viewing the configuration file
