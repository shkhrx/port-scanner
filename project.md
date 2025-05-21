%%{init: {'themeVariables': { 'fontSize': '12px' }}}%%
graph TD
    A[Start Scan] --> B[Frontend Input]
    B --> C{Valid Target & Ports?}
    C -- Yes --> D[Go Backend: Scan with Goroutines]
    C -- No --> B
    D --> E[Check Ports on Target]
    E --> F{Status}
    F -- Open --> G[Open]
    F -- Closed/Filtered --> H[Closed/Filtered]
    G & H --> I[GeoIP + Store Results]
    I --> J[Show on UI]
    J --> K{User Options}
    K -- Export --> L[Download JSON/CSV]
    K -- New Scan --> B
