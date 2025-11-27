Factory
    Implemented in "internal/games/factory.go"
    Factory was implemented to allow for new games to be added without the need of
    modifying server code. This demonstrates the open/closed principle, all that is
    needed is to add new cases

Singleton
    Implemented in "internal/manager/manager.go" and "internal/dictionary/dictionary.go"
    Singleton was implemented for consistency and to avoid reloading large datasets,
    ensuring thread-safe and global access to shared states.

Adaptor
    Implemented in "internal/dictionary/adapter/adapter.go"
    Adapter was implemented to allow the use of alternative dictionary sources without
    changing game logic. This demonstrates the loose coupling principle decreasing
    dependencies and increasing system flexibility