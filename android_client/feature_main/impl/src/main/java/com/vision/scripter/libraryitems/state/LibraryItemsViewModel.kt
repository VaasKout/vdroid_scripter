package com.vision.scripter.libraryitems.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.libraryitems.ui.LibraryItemsUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
internal class LibraryItemsViewModel @Inject constructor(
    private val libraryItemsInteractor: LibraryItemsInteractor,
) : ViewModel(), LibraryItemsUiStateHolder by libraryItemsInteractor {

    override fun onCleared() {
        super.onCleared()
        libraryItemsInteractor.clear()
    }
}
