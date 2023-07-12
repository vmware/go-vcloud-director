* Added RDE Defined Interface Behaviors support with methods `DefinedInterface.AddBehavior`, `DefinedInterface.GetAllBehaviors`,
  `DefinedInterface.GetBehaviorById` `DefinedInterface.GetBehaviorByName`, `DefinedInterface.UpdateBehavior` and
  `DefinedInterface.DeleteBehavior` [GH-584]
* Added RDE Defined Entity Type Behaviors support with methods `DefinedEntityType.GetAllBehaviors`,
  `DefinedEntityType.GetBehaviorById` `DefinedEntityType.GetBehaviorByName`, `DefinedEntityType.UpdateBehaviorOverride` and
  `DefinedEntityType.DeleteBehaviorOverride` [GH-584]
* Added RDE Defined Entity Type Behavior Access Controls support with methods `DefinedEntityType.GetAllBehaviorsAccessControls` and 
  `DefinedEntityType.SetBehaviorAccessControls` [GH-584]
* Added method to invoke Behaviors on Defined Entities `DefinedEntity.InvokeBehavior` and `DefinedEntity.InvokeBehaviorAndMarshal` [GH-584]
