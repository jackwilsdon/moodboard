module Main exposing (main)

import Browser
import Html exposing (Html)
import Html.Attributes as Attributes


type alias Position =
    { x : Float, y : Float }


type DragState
    = Dragging { item : String, grab : Position, mouse : Position }


type alias Model =
    { items : List String }


type Msg
    = NoOp


init : Model
init =
    { items = List.range 0 10 |> List.map String.fromInt }


update : Msg -> Model -> Model
update msg model =
    case msg of
        NoOp ->
            model


view : Model -> Html Msg
view model =
    Html.div []
        (List.map
            (\item ->
                Html.div
                    [ Attributes.class "item" ]
                    [ Html.text ("Item: " ++ item) ]
            )
            model.items
        )


main : Program () Model Msg
main =
    Browser.sandbox
        { init = init
        , view = view
        , update = update
        }
