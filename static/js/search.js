$(document).ready(function () {
    // Search functionality
    $('#searchInput').on('input', function () {
        var query = $(this).val().trim();
        if (query.length === 0) {
            // location.reload();
            $("#saccoContent").show();
            $('#searchResultsTableContent').hide();
            $("#specificSaccoContent").hide();
            // Clear the details when the search input is cleared
            clearDetails();
        }
    });

    // Focuses on search input when searching
    $(document).ready(function() {
        $("#button").click(function() {
          $("#searchform").slideToggle("fast");
          $('#search').focus();
        });
    });
    
    $('#searchForm').submit(function (event) {
        event.preventDefault();
        var query = $('#searchInput').val().trim();

        if (!query) {
            alert('Please enter a search query.');
            return;
        }

        // Hide specific SACCO content when performing a search
        $("#specificSaccoContent").hide();

        // Hide and clear suggestions
        hideSuggestions();

        $.ajax({
            type: 'GET',
            url: '/search',
            data: { q: query },
            dataType: 'json',
            success: function (response) {
                console.log("Search Results:", response);
                $("#saccoContent").hide();
                $('#searchResultsTableContent').show();
                var searchResultsContainer = $('#searchResultsTableContent');

                if (searchResultsContainer.length === 0) {
                    searchResultsContainer = $('<table id="searchResultsTableContent" class="table"><thead><tr></tr></thead><tbody></tbody></table>');
                    $('body').append(searchResultsContainer);
                }

                var tableBody = searchResultsContainer.find('tbody');
                var tableHeaderRow = searchResultsContainer.find('thead tr');
                tableBody.empty();
                tableHeaderRow.empty();
                
                // Append table headings
                tableHeaderRow.append('<th>Result Type</th>');
                if (response.Drivers && response.Drivers.length > 0) {
                    // Append headings for drivers
                    tableHeaderRow.append('<th>Name</th>');
                    tableHeaderRow.append('<th>ID Number</th>');
                    tableHeaderRow.append('<th>Contact</th>');
                    tableHeaderRow.append('<th>Sacco</th>');
                } else if (response.Cars && response.Cars.length > 0) {
                    // Append headings for cars
                    tableHeaderRow.append('<th>Number Plate</th>');
                    tableHeaderRow.append('<th>Make</th>');
                    tableHeaderRow.append('<th>Model</th>');
                    tableHeaderRow.append('<th>Sacco</th>');
                    tableHeaderRow.append('<th>Trips Per Day</th>');
                } else if (response.Managers && response.Managers.length > 0) {
                    // Append headings for managers
                    tableHeaderRow.append('<th>Name</th>');
                    tableHeaderRow.append('<th>Contact</th>');
                    tableHeaderRow.append('<th>Sacco</th>');
                } else if (response.Saccos && response.Saccos.length > 0) {
                    // Append headings for saccos
                    tableHeaderRow.append('<th>Sacco Name</th>');
                    tableHeaderRow.append('<th>Manager</th>');
                    tableHeaderRow.append('<th>Contact</th>');
                }

                if (response && typeof response === 'object') {
                    if (response.Drivers && Array.isArray(response.Drivers) && response.Drivers.length > 0) {
                        $.each(response.Drivers, function (index, driver) {
                            var row = '<tr>' +
                                '<td>Driver</td>' +
                                '<td>' + driver.Name + '</td>' +
                                '<td>' + driver.IDNumber + '</td>' +
                                '<td>' + driver.Contact + '</td>' +
                                '<td>' + driver.SaccoName + '</td>' +
                                '</tr>';
                            tableBody.append(row);
                        });
                    }

                    if (response.Cars && Array.isArray(response.Cars) && response.Cars.length > 0) {
                        $.each(response.Cars, function (index, car) {
                            var row = '<tr>' +
                                '<td>Car</td>' +
                                '<td>' + car.NumberPlate + '</td>' +
                                '<td>' + car.Make + '</td>' +
                                '<td>' + car.Model + '</td>' +
                                '<td>' + car.SaccoName + '</td>' +
                                '<td>' + car.Trips + '</td>' +
                                '</tr>';
                            tableBody.append(row);
                        });
                    }

                    if (response.Managers && Array.isArray(response.Managers) && response.Managers.length > 0) {
                        $.each(response.Managers, function (index, manager) {
                            var row = '<tr>' +
                                '<td>Manager</td>' +
                                '<td>' + manager.Manager + '</td>' +
                                '<td>' + manager.Contact + '</td>' +
                                '<td>' + manager.SaccoName + '</td>' +
                                '</tr>';
                            tableBody.append(row);
                        });
                    }
                    
                    if (response && Array.isArray(response.Saccos) && response.Saccos.length > 0) {
                        $.each(response.Saccos, function (index, sacco) {
                            var row = '<tr>' +
                                '<td>Sacco</td>' +
                                '<td>' + sacco.SaccoName + '</td>' +
                                '<td>' + sacco.Manager + '</td>' +
                                '<td>' + sacco.Contact + '</td>' +
                                '</tr>';
                            tableBody.append(row);
                        });
                    }
                } else {
                    console.error('Invalid response:', response);
                }
            },
            error: function (xhr, status, error) {
                console.error('Error:', error);
            }
        });
    });


    // Search suggestions
    $('#searchInput').on('input', function () {
        var query = $(this).val().trim();
        if (query.length > 0) {
            $.ajax({
                type: 'GET',
                url: '/search-suggestions',
                data: { q: query },
                dataType: 'json',
                success: function (response) {
                    displaySuggestions(response);

                    // Hide sacco content when searching
                    $("#saccoContent").hide();
                },
                error: function (xhr, status, error) {
                    console.error('Error fetching search suggestions:', error);
                }
            });
        } else {
            $("#saccoContent").show();
            $('#searchSuggestions').empty();
        }
    });

    // Function to display search suggestions
    function displaySuggestions(suggestions) {
        var suggestionList = $('#searchSuggestions');
        suggestionList.empty();
        if (suggestions && suggestions.length > 0) {
            $.each(suggestions, function (index, suggestion) {
                suggestionList.append('<div class="suggestion-link">' + suggestion + '</div>');
            });

            // Show the search suggestions container
            suggestionList.show();

            // Clear displayed details
            clearDetails();

            // Hide SACCO link details when displaying suggestions
            $("#specificSaccoContent").hide();

        } else {
            suggestionList.append('<div>No suggestions found</div>');
        }
    }

    // Hiding suggestions after clicking a suggestion
    function hideSuggestions() {
        $('#searchSuggestions').empty(); // Clear the suggestion list
        $('#searchSuggestions').hide(); // Hide the suggestion list container
    }

    // Fetch details of selected suggestion
    $(document).on("click", ".suggestion-link", function () {
        var suggestion = $(this).text().trim();
        fetchDetails(suggestion); // Fetch details for the clicked suggestion
        hideSuggestions(); // Hide suggestions after clicking
    });

    function fetchDetails(suggestion) {
        $.ajax({
            type: 'GET',
            url: '/get-details',
            data: { suggestion: suggestion },
            dataType: 'json',
            success: function (response) {
                console.log("Details:", response);
                displayDetails(response);
            },
            error: function (xhr, status, error) {
                console.error('Error fetching details:', error);
            }
        });
    }

    // Display details of selected suggestion
    function displayDetails(details) {
        var detailsContainer = $('#detailsContainer');
        detailsContainer.empty();
    
        if (details && Object.keys(details).length > 0) {
            var table = $('<table>').addClass('details-table');
            var tableBody = $('<tbody>');
    
            // Row for headings
            var headingsRow = $('<tr>');
            $.each(details, function (key, value) {
                if (typeof value === 'object') {
                    // If the value is an object, iterate over its properties and add them as headings
                    $.each(value, function (subKey, _) {
                        headingsRow.append($('<th>').text(subKey));
                    });
                } else {
                    headingsRow.append($('<th>').text(key));
                }
            });
            tableBody.append(headingsRow);
    
            // Row for details
            var detailsRow = $('<tr>');
            $.each(details, function (_, value) {
                if (typeof value === 'object') {
                    // If the value is an object, iterate over its properties and add them as row data
                    $.each(value, function (_, subValue) {
                        detailsRow.append($('<td>').text(subValue));
                    });
                } else {
                    detailsRow.append($('<td>').text(value));
                }
            });
            tableBody.append(detailsRow);
    
            table.append(tableBody);
            detailsContainer.append(table);
        } else {
            detailsContainer.append('<div>No details found</div>');
        }
    }

    // Hiding display results after clicking a suggestion
    function clearDetails() {
        $('#detailsContainer').empty(); // Clear the display list
    }

    // Function to hide page details table
    function hideHomeDetailsTable() {
        $(".container").hide();
    }

    // Function to show page details table
    function showHomeDetailsTable() {
        $(".container").show();
    }

    // Event listener for search input to hide pages' details
    $('#searchInput').on('input', function () {
        var query = $(this).val().trim();
        if (query.length > 0) {
            // Hide details table when search is performed
            hideHomeDetailsTable();
        } else {
            // Show details table when search input is empty
            showHomeDetailsTable();
        }
    });
});