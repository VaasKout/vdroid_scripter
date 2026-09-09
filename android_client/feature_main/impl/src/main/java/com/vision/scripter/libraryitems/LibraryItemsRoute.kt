package com.vision.scripter.libraryitems

import com.vision.scripter.library.state.LibraryType

internal const val LibraryTypeArg = "type"
internal const val LibraryItemsRoute = "library"
internal const val LibraryItemsRouteWithArgs = "$LibraryItemsRoute/{$LibraryTypeArg}"

internal fun libraryItemsRoute(type: LibraryType): String = "$LibraryItemsRoute/${type.name}"
