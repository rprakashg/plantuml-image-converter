#plantuml-image-conversion
Converts any plant UML diagram to PNG image. For more info [see](http://www.plantuml.com)

export PROJECT_ID = "set it to your GCP project"

## Building
Run command below in shell


```

gcloud builds submit . --config=build.yaml

```

## Running
Run command below to deploy and run it in Cloud Run

```
gcloud beta run deploy gcr.io/$PROJECT_ID/plantuml-image-conversion
```

## Tesitng
Save the sample UML diagram below on to a file named demo.uml

```
@startuml
Alice -> Bob: Authentication Request
Bob --> Alice: Authentication Response

Alice -> Bob: Another authentication Request
Alice <-- Bob: another authentication Response
@enduml

```

CURL

```

``