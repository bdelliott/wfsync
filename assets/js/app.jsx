class SyncItem extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            syncStatus: "Unknown"
        };
    }

    componentDidMount() {
        // fetch state of component
    }

    render() {
        return (
            <li>sync state of {this.props.name}...TBD</li>
        )
    }

}

class Synchronizer extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            status: "initializing"
        }
    }

    componentDidMount() {
        this.syncStatus();
    }

    render() {

        return (
            <ul>
                <SyncItem name="Nokia"/>
                <SyncItem name="FatSecret"/>
            </ul>
        )
    }

    updateSyncStatus() {
        // update status values
        alert('update');
    }

    syncStatus() {
        // call server to update sync progress and check statuses
        var url = window.location.href + "syncStatus";
        fetch(url).then(this.updateSyncStatus);
    }

  
}

ReactDOM.render(
      <Synchronizer/>,
      document.getElementById('root')
);
