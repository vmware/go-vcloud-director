* Add `AdminOrg` methods `CreateCatalogFromSubscriptionAsync` and `CreateCatalogFromSubscription` to create a
  subscribed catalog [GH-511]
* Add method `AdminCatalog.FullSubscriptionUrl` to return the subscription URL of a published catalog [GH-511]
* Add method `AdminCatalog.WaitForTasks` to wait for catalog tasks to complete [GH-511]
* Add method `AdminCatalog.UpdateSubscriptionParams` to modify the terms of an existing subscription [GH-511]
* Add methods `Catalog.QueryTaskList` and `AdminCatalog.QueryTaskList` to retrieve the tasks associated with a catalog [GH-511]
* Add function `IsValidUrl` to determine if a URL is valid [GH-511]
* Add `AdminCatalog` methods `Sync` and `LaunchSync` to synchronise a subscribed catalog [GH-511]
* Add method `AdminCatalog.GetCatalogHref` to retrieve the HREF of a regular catalog [GN-511]
* Add `AdminCatalog` methods `QueryCatalogItemList`, `QueryVappTemplateList`, and `QueryMediaList` to retrieve lists of
  dependent items [GH-511]
* Add  `AdminCatalog` methods `LaunchSynchronisationVappTemplates`, `LaunchSynchronisationAllVappTemplates`,
  `LaunchSynchronisationMediaItems`, and `LaunchSynchronisationAllMediaItems` to start synchronisation of dependent
  items [GH-511]
* Add `AdminCatalog` methods `GetCatalogItemByHref` and `QueryCatalogItem` to retrieve a single Catalog Item [GH-511]
* Add method `CatalogItem.LaunchSync` to start synchronisation of a catalog item [GH-511]
* Add method `CatalogItem.Refresh` to get fresh contents for a catalog item [GH-511]
* Add function `WaitResource` to wait for tasks associated to a gioven resource [GH-511]
* Add function `MinimalShowTask` to display task progress with minimal info [GH-511]
* Add functions `ResourceInProgress` and `ResourceComplete` to check on task activity for a given entity [GH-511]
* Add functions `SkimTasksList`, `SkimTasksListMonitor`, `WaitTaskListCompletion`, `WaitTaskListCompletionMonitor` to
  process lists of tasks and lists of task IDs [GH-511]
* Add `Client` methods `GetTaskByHREF` and `GetTaskById` to retrieve individual tasks [GH-511]
* Implement `QueryItem` for `Task` and `AdminTask` (`GetHref`, `GetName`, `GetType`, `GetParentId`, `GetParentName`, `GetMetadataValue`, `GetDate`) [GH-511]
