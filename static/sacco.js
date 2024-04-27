$(document).ready(function () {
    // Initialize Bootstrap dropdowns
    $('.dropdown-toggle').dropdown();

    // Submit SACCO form
    $("#submitSaccoBtn").click(function () {
        var saccoName = $("#saccoName").val();
        var jsonData = {
            "saccoName": saccoName
        };

        console.log("JSON Data: ", jsonData);

        $.ajax({
            type: "POST",
            url: "/add-sacco",
            data: JSON.stringify(jsonData),
            contentType: "application/json",
            success: function () {
                console.log("Sacco added successfully");
                location.reload();
            },
            error: function (xhr, status, error) {
                console.error('Failed to add Sacco: ' + error);
            },
            complete: function () {
                $("#addSaccoModal").modal("hide");
            }
        });
    });

    // Search functionality
    $('#searchInput').on('input', function () {
        var query = $(this).val().trim();
        if (query.length === 0) {
            $("#saccoContent").show();
            $('#searchResultsTableContent').hide();
        }
    });
    
    $('#searchForm').submit(function (event) {
        event.preventDefault();
        var query = $('#searchInput').val().trim();

        if (!query) {
            alert('Please enter a search query.');
            return;
        }

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
                    tableHeaderRow.append('<th>Name</th>');
                    tableHeaderRow.append('<th>ID Number</th>');
                    tableHeaderRow.append('<th>Contact</th>');
                    tableHeaderRow.append('<th>Sacco</th>');
                } else if (response.Cars && response.Cars.length > 0) {
                    tableHeaderRow.append('<th>Number Plate</th>');
                    tableHeaderRow.append('<th>Make</th>');
                    tableHeaderRow.append('<th>Model</th>');
                    tableHeaderRow.append('<th>Sacco</th>');
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
                                '</tr>';
                            tableBody.append(row);
                        });
                    }

                    if (response.Saccos && Array.isArray(response.Saccos) && response.Saccos.length > 0) {
                        $.each(response.Saccos, function (index, sacco) {
                            tableBody.append('<tr><td>Sacco</td><td>' + sacco.SaccoName + '</td><td></td><td></td><td></td></tr>');
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
                },
                error: function (xhr, status, error) {
                    console.error('Error fetching search suggestions:', error);
                }
            });
        } else {
            $('#searchSuggestions').empty();
        }
    });

    function displaySuggestions(suggestions) {
        var suggestionList = $('#searchSuggestions');
        suggestionList.empty();
        if (suggestions && suggestions.length > 0) {
            $.each(suggestions, function (index, suggestion) {
                suggestionList.append('<div class="suggestion-link">' + suggestion + '</div>');
            });
        } else {
            suggestionList.append('<div>No suggestions found</div>');
        }
    }

    // Fetch details of selected suggestion
    $(document).on("click", ".suggestion-link", function () {
        var suggestion = $(this).text().trim();
        fetchDetails(suggestion);
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

    function displayDetails(details) {
        var detailsContainer = $('#detailsContainer');
        detailsContainer.empty();

        if (details && Object.keys(details).length > 0) {
            $.each(details, function (key, value) {
                detailsContainer.append('<div><strong>' + key + ':</strong> ' + value + '</div>');
            });
        } else {
            detailsContainer.append('<div>No details found</div>');
        }
    }

    // Event listener for SACCO link click
    $(document).ready(function () {
        // Push state when SACCO link is clicked
        $(".sacco-link").click(function (e) {
            e.preventDefault();
            var saccoID = $(this).data("saccoid");
            var saccoName = $(this).text().trim();

            // Push state to history
            history.pushState({ page: "specificSacco", saccoID: saccoID }, saccoName, '?sacco=' + saccoID);

            // Fetch SACCO details and show specific SACCO content
            fetchSaccoDetails(saccoID, saccoName);
        });

        // Listen for popstate event (browser back/forward buttons)
        $(window).on('popstate', function (e) {
            var state = e.originalEvent.state;
            if (state && state.page === 'specificSacco') {
                fetchSaccoDetails(state.saccoID, document.title); // Fetch SACCO details based on state
            } else {
                // If not specific SACCO page, show default SACCO content
                showDefaultSaccoContent();
            }
        });

        // Function to fetch SACCO details and show specific SACCO content
        function fetchSaccoDetails(saccoID, saccoName) {
            $.ajax({
                type: "GET",
                url: "/get-cars-and-drivers-routes?saccoID=" + saccoID,
                success: function (data) {
                    console.log("Cars, drivers, and routes data retrieved successfully:", data);
                    $("#specificSaccoHeader h1").text(saccoName);
                    renderCars(data.Cars);
                    renderDrivers(data.Drivers);
                    $("#saccoContent").hide();
                    $("#specificSaccoContent").show();
                    $("#specificSaccoDetails").show();
                },
                error: function (xhr, status, error) {
                    console.error("Error fetching cars and drivers:", error);
                    alert("Failed to fetch cars and drivers. Please try again.");
                }
            });
        }

        // Function to show default SACCO content
        function showDefaultSaccoContent() {
            $("#saccoContent").show();
            $("#specificSaccoContent").hide();
            $("#specificSaccoDetails").hide();
        }

        // Initial check for state on page load
        if (window.location.search.indexOf('sacco=') === -1) {
            // If no specific SACCO query parameter is present, show default SACCO content
            showDefaultSaccoContent();
        }
    });

    function renderCars(cars) {
        var carsTableBody = $("#carsTable tbody");
        carsTableBody.empty();

        if (cars && cars.length > 0) {
            $.each(cars, function (index, car) {
                var row = "<tr><td>" + car.NumberPlate + "</td><td>" + car.Make + "</td><td>" + car.Model + "</td></tr>";
                carsTableBody.append(row);
            });
            $("#carsContainer").show(); // Show cars container if there are cars
        } else {
            // Show message if no cars available
            carsTableBody.html('<tr><td colspan="3">No cars available</td></tr>');
            $("#carsContainer").show(); // Show cars container with message
        }
    }

    function renderDrivers(drivers) {
        var driversTableBody = $("#driversTable tbody");
        driversTableBody.empty();

        if (drivers && drivers.length > 0) {
            $.each(drivers, function (index, driver) {
                var row = "<tr><td>" + driver.Name + "</td><td>" + driver.IDNumber + "</td><td>" + driver.Contact + "</td></tr>";
                driversTableBody.append(row);
            });
            $("#driversContainer").show(); // Show drivers container if there are drivers
        } else {
            // Show message if no drivers available
            driversTableBody.html('<tr><td colspan="3">No drivers available</td></tr>');
            $("#driversContainer").show(); // Show drivers container with message
        }
    }

    // function renderRoutes(routes) {
    //     var routesTableBody = $("#routesTable tbody");
    //     routesTableBody.empty();

    //     if (routes && routes.length > 0) {
    //         $.each(routes, function (index, route) {
    //             // var row = "<tr><td>" + route.Checkpoint + "</td></tr>";
    //             var row = "<tr><td>" + route.ID + "</td><td>" + route.Checkpoints.join(', ') + "</td></tr>";
    //             routesTableBody.append(row);
    //         });
    //         $("#routesContainer").show();
    //     } else {
    //         $("#routesContainer").hide();
    //     }
    // }
});